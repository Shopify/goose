package genmain_test

import (
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"

	"github.com/Shopify/goose/genmain"
)

type testComponent struct {
	tomb    tomb.Tomb
	started chan struct{}
}

func newTestComponent() *testComponent {
	return &testComponent{
		started: make(chan struct{}, 1),
	}
}

func (c *testComponent) Tomb() *tomb.Tomb {
	return &c.tomb
}

func (c *testComponent) Run() error {
	c.started <- struct{}{}
	<-c.tomb.Dying()
	return c.tomb.Err()
}

type hangingComponent struct {
	tomb     tomb.Tomb
	hangTime time.Duration
}

func newHangingComponent(hangTime time.Duration) *hangingComponent {
	return &hangingComponent{
		hangTime: hangTime,
	}
}

func (c *hangingComponent) Tomb() *tomb.Tomb {
	return &c.tomb
}

func (c *hangingComponent) Run() error {
	<-time.After(c.hangTime)
	return c.tomb.Err()
}

func TestKillPropagatesErrorToComponents(t *testing.T) {
	expectedErr := errors.New("expected")
	_, actualErr := runAndPerform(t, func(m *genmain.Main, c *testComponent) {
		m.Kill(expectedErr)
	})
	assert.Equal(t, expectedErr, actualErr)
}

func TestSignalErrorPropagatesToComponents(t *testing.T) {
	_, actualErr := runAndPerform(t, func(_m *genmain.Main, c *testComponent) {
		if p, err := os.FindProcess(os.Getpid()); err != nil {
			t.Fatal("couldn't find process")
		} else {
			p.Signal(syscall.SIGINT)
		}
	})
	assert.Equal(t, "received signal: interrupt", actualErr.Error())
}

func TestKillOnlyUsesFirstError(t *testing.T) {
	expectedErr := errors.New("expected")
	_, actualErr := runAndPerform(t, func(m *genmain.Main, c *testComponent) {
		m.Kill(expectedErr)
		m.Kill(errors.New("not expected"))
	})
	assert.Equal(t, expectedErr, actualErr)
}

func TestRunAndWaitReturnsErrorIfAlreadyRan(t *testing.T) {
	main, _ := runAndPerform(t, func(_m *genmain.Main, c *testComponent) {
		c.tomb.Kill(nil)
	})
	assert.Equal(t, genmain.ErrCanOnlyRunOnce, main.RunAndWait())
}

func TestShutdownDeadline(t *testing.T) {
	deadline := 500 * time.Millisecond

	main := genmain.New(
		// These 3 will shut down in time, but if they are killed sequentially, they will add up to more than the deadline
		newHangingComponent(deadline*2/3),
		newHangingComponent(deadline*2/3),
		newHangingComponent(deadline*2/3),

		// This won't shut down in time, but would if the deadline is not being executed.
		newHangingComponent(deadline*3/2),
	)
	main.SetShutdownDeadline(deadline)

	start := time.Now()
	main.Kill(nil)
	main.RunAndWait()
	shutdownTime := time.Since(start)

	assert.True(t, shutdownTime >= deadline, "genmain should wait until deadline before killing a hanging component")
	assert.True(t, shutdownTime < deadline*3/2, "genmain should kill a hanging component after the deadline")
}

func runAndPerform(t *testing.T, perform func(*genmain.Main, *testComponent)) (*genmain.Main, error) {
	component := newTestComponent()

	completed := make(chan struct{})
	main := genmain.New(component)

	deadline := 1 * time.Second
	main.SetShutdownDeadline(deadline)

	go (func() {
		main.RunAndWait()
		completed <- struct{}{}
	})()

	select {
	case <-component.started:
	case <-time.After(deadline / 3):
		t.Fatal("expected component to be running")
	}

	perform(&main, component)

	select {
	case <-completed:
	case <-time.After(deadline / 3):
		t.Fatal("expected main to terminate")
	}

	return &main, component.tomb.Err()
}
