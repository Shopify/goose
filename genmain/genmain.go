package genmain

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"

	"gopkg.in/tomb.v2"

	"github.com/Shopify/goose/logger"
	"github.com/Shopify/goose/metrics"
	"github.com/Shopify/goose/safely"
	"github.com/Shopify/goose/statsd"
)

var log = logger.New("genmain")

const defaultShutdownDeadline = 3 * time.Second

var (
	// ErrCanOnlyRunOnce is returned when `RunAndWait` is called after already
	// being called.
	ErrCanOnlyRunOnce = errors.New("can only run once")

	// ErrShutdownRequested can be used as a reason for `Kill` that indicates no
	// error has occurred, just that the components should gracefully exit.
	ErrShutdownRequested = errors.New("shutdown requested")
)

// Main represents a collection of components whose lifecycles are tied together.
type Main struct {
	components       []Component
	shutdownDeadline time.Duration

	l   sync.Mutex
	ran bool
}

// New creates a new `Main`
func New(components ...Component) Main {
	return Main{
		components:       components,
		shutdownDeadline: defaultShutdownDeadline,
	}
}

// SignalError is an `error` returned when `Main` exits due to a signal.
type SignalError struct {
	signal os.Signal
}

func (s *SignalError) Error() string {
	return fmt.Sprintf("received signal: %v", s.signal)
}

func waitAny(components []Component, deadline <-chan time.Time) bool {
	cases := make([]reflect.SelectCase, len(components)+1)
	for i, component := range components {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(component.Tomb().Dead())}
	}
	cases[len(components)] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(deadline)}
	chosen, _, _ := reflect.Select(cases)

	return chosen < len(components)
}

func dependencies(component Component) []Component {
	if c, ok := component.(ComponentWithDependencies); ok {
		return c.Dependencies()
	}

	return nil
}

func inverseDependencies(components []Component) map[Component][]Component {
	inverse := make(map[Component][]Component, len(components))
	for _, component := range components {
		for _, dep := range dependencies(component) {
			inverse[dep] = append(inverse[dep], component)
		}
	}
	return inverse
}

func isReadyToKill(inverse map[Component][]Component, component Component) bool {
	if deps, ok := inverse[component]; ok {
		for _, dep := range deps {
			if !isDead(dep.Tomb()) {
				return false
			}
		}
	}

	return true
}

func isDead(t *tomb.Tomb) bool {
	select {
	case <-t.Dead():
		return true
	default:
		return false
	}
}

// Kill will terminate all running components with a given reason
func (m *Main) Kill(reason error) {
	// Acquire the lock to ensure the first call's `err` is the one that all components
	// receive, and not some mix if this function was called concurrently.
	m.l.Lock()
	defer m.l.Unlock()

	log(nil, reason).Info("shutting down")
	shutdownStart := time.Now()
	deadline := time.After(m.shutdownDeadline)

	ctx := context.Background()
	alive := m.components
	inverseDeps := inverseDependencies(m.components)

	for len(alive) > 0 {
		var killed []Component

		for _, component := range alive {
			if isReadyToKill(inverseDeps, component) {
				component.Tomb().Kill(reason)
				killed = append(killed, component)
			}
		}

		if !waitAny(killed, deadline) {
			for _, component := range alive {
				metrics.GenMainShutdown.Duration(ctx, time.Since(shutdownStart), statsd.Tags{
					"success":       false,
					"deadline":      m.shutdownDeadline,
					"mainComponent": componentName(component),
				})
				log(ctx, component.Tomb().Err()).
					WithField("mainComponent", componentName(component)).
					Error("component took too long to shut down")
			}
			return
		}

		i := 0
		for _, component := range alive {
			if isDead(component.Tomb()) {
				metrics.GenMainShutdown.Duration(ctx, time.Since(shutdownStart), statsd.Tags{
					"success":       true,
					"deadline":      m.shutdownDeadline,
					"mainComponent": componentName(component),
				})
				log(ctx, component.Tomb().Err()).
					WithField("mainComponent", componentName(component)).
					Debug("component shutdown notification")
			} else {
				alive[i] = component
				i++
			}
		}
		alive = alive[0:i]
	}
}

func (m *Main) SetShutdownDeadline(d time.Duration) {
	m.shutdownDeadline = d
}

// RunAndWait starts all components in this `Main`.
//
// `RunAndWait` will also listen to SIGINT and SIGTERM to do graceful shutdowns of all
// components it manages. It should only be called once, and returns `ErrCanOnlyRunOnce`
// if called more than once.
func (m *Main) RunAndWait() error {
	ctx := context.Background()
	defer metrics.GenMainRun.StartTimer(ctx).Finish()

	m.l.Lock()
	if m.ran {
		m.l.Unlock()
		return ErrCanOnlyRunOnce
	}

	m.ran = true
	m.l.Unlock()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	shutdown := make(chan error, len(m.components)+1)
	safely.Go(func() {
		sig := <-sigs
		go func() {
			<-sigs
			log(nil, nil).Fatal("received signal again: terminating immediately")
		}()

		shutdown <- &SignalError{sig}
	})

	for _, c := range m.components {
		safely.Run(c)

		go func(comp Component) {
			<-comp.Tomb().Dead()
			err := comp.Tomb().Err()
			log(nil, err).
				WithField("mainComponent", componentName(comp)).
				Info("process exited")
			shutdown <- err
		}(c)
	}

	reason := <-shutdown
	m.Kill(reason)

	log(nil, nil).Debug("final shutdown message")

	return reason
}

func componentName(comp Component) string {
	compType := fmt.Sprintf("%T", comp)
	return strings.TrimPrefix(compType, "*")
}
