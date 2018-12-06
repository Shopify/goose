package statetracker

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStateTracker_Set(t *testing.T) {
	tracker := New(false)

	// Default
	assert.False(t, tracker.Get().(bool))
	// Very small for now, but should still be non-zero
	assert.True(t, tracker.Duration(false, true) > 0)

	// Still false
	tracker.Set(false)
	assert.False(t, tracker.Get().(bool))

	// Now true
	tracker.Set(true)
	assert.True(t, tracker.Get().(bool))
	assert.True(t, tracker.Duration(false, true) == 0)
}

func TestStateTracker_Wait(t *testing.T) {
	tracker := New(false)

	waiting := make(chan struct{}, 1)
	ready := make(chan struct{}, 1)
	go func() {
		waiting <- struct{}{}
		assert.True(t, tracker.Wait(context.Background(), true))
		ready <- struct{}{}
	}()

	assert.True(t, len(ready) == 0, "should not yet be ready")

	tracker.Set(false)
	assert.True(t, len(ready) == 0, "should not yet be ready")

	select {
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "should be waiting by now")
	case <-waiting:
	}
	tracker.Set(true)

	select {
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "should be ready by now")
	case <-ready:
	}
}

func TestStateTracker_Listener(t *testing.T) {
	var runCount uint32
	dead := make(chan struct{}, 1)
	cond := sync.NewCond(&sync.Mutex{})

	tracker := New(false)
	listener := tracker.NewListener()
	go (func() {
		for n := range listener {
			c := atomic.AddUint32(&runCount, 1)

			switch c {
			case 1:
				assert.Equal(t, false, n.Old)
				assert.Equal(t, false, n.New)
				assert.True(t, n.SinceOld > 0)
				assert.Equal(t, n.SinceOld, n.SinceLastNew)
				assert.False(t, n.IsInitial())
			case 2:
				assert.Equal(t, false, n.Old)
				assert.Equal(t, true, n.New)
				assert.True(t, n.SinceOld > 0)
				assert.True(t, n.SinceLastNew == 0)
				assert.True(t, n.IsInitial())
			case 3:
				assert.Equal(t, true, n.Old)
				assert.Equal(t, false, n.New)
				assert.True(t, n.SinceOld > 0)
				assert.True(t, n.SinceLastNew > 0)
				assert.NotEqual(t, n.SinceOld, n.SinceLastNew)
				assert.False(t, n.IsInitial())
			default:
				assert.Fail(t, "should not be invoked more than 3 times")
			}
			cond.Signal()
		}
		close(dead)
	})()

	assert.Equal(t, uint32(0), atomic.LoadUint32(&runCount))

	tracker.Set(false)
	cond.L.Lock()
	cond.Wait()
	assert.Equal(t, uint32(1), atomic.LoadUint32(&runCount))
	cond.L.Unlock()

	tracker.Set(true)
	cond.L.Lock()
	cond.Wait()
	assert.Equal(t, uint32(2), atomic.LoadUint32(&runCount))
	cond.L.Unlock()

	tracker.Set(false)
	cond.L.Lock()
	cond.Wait()
	assert.Equal(t, uint32(3), atomic.LoadUint32(&runCount))
	cond.L.Unlock()

	tracker.RemoveListener(listener)
	select {
	case <-time.After(100 * time.Millisecond):
		assert.Fail(t, "should be dead by now")
	case <-dead:
	}

	// Check it doesn't crash
	tracker.Set(true)
}
