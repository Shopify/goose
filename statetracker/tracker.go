package statetracker

import (
	"context"
	"sync"
	"time"

	"github.com/Shopify/goose/v2/logger"
)

var log = logger.New("stateTracker")

type StateTracker interface {
	Get() interface{}
	Set(interface{}) bool

	// The time that the StateTracker has spent being "previous"
	// If already "desired", returns 0
	Duration(previous interface{}, desired interface{}) time.Duration
	Wait(ctx context.Context, desired interface{}) bool

	NewListener() <-chan *Notification
	RemoveListener(listener <-chan *Notification)
}

type stateTracker struct {
	cond      *sync.Cond
	value     interface{}
	lastSet   map[interface{}]time.Time
	listeners []chan *Notification
}

func New(initial interface{}) StateTracker {
	return &stateTracker{
		cond:  sync.NewCond(&sync.RWMutex{}),
		value: initial,
		lastSet: map[interface{}]time.Time{
			initial: time.Now(),
		},
	}
}

func (t *stateTracker) Duration(previous interface{}, desired interface{}) time.Duration {
	t.cond.L.(*sync.RWMutex).RLock()
	defer t.cond.L.(*sync.RWMutex).RUnlock()

	if t.value == desired {
		return 0
	}

	return time.Since(t.lastSet[previous])
}

func (t *stateTracker) Get() interface{} {
	t.cond.L.(*sync.RWMutex).RLock()
	defer t.cond.L.(*sync.RWMutex).RUnlock()

	return t.value
}

// Set changes the underlying value, triggers callbacks, and returns whether the value has changed
func (t *stateTracker) Set(val interface{}) bool {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	old := t.value
	now := time.Now()

	sinceOld := now.Sub(t.lastSet[old])
	sinceLastNew := time.Duration(0)
	if lastNew, ok := t.lastSet[val]; ok {
		sinceLastNew = now.Sub(lastNew)
	}

	t.value = val
	t.lastSet[val] = now

	t.cond.Broadcast()

	if len(t.listeners) > 0 {
		n := &Notification{
			New:          val,
			Old:          old,
			SinceOld:     sinceOld,
			SinceLastNew: sinceLastNew,
		}
		for _, listener := range t.listeners {
			listener <- n
		}
	}

	return old != val
}

// Wait until underlying state is the desired one, returning whether it had to wait or not.
func (t *stateTracker) Wait(ctx context.Context, desired interface{}) bool {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	waited := false

	if t.value != desired {
		waited = true
		log(ctx, nil).
			WithField("current", t.value).
			WithField("desired", desired).
			Info("waiting for a state change")
	}

	// Can be broadcasted many times, wait for the right desired
	for t.value != desired {
		t.cond.Wait()
	}

	return waited
}

func (t *stateTracker) NewListener() <-chan *Notification {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	l := make(chan *Notification, 1)
	t.listeners = append(t.listeners, l)
	return l
}

func (t *stateTracker) RemoveListener(listener <-chan *Notification) {
	t.cond.L.Lock()
	defer t.cond.L.Unlock()

	for i, l := range t.listeners {
		if l == listener {
			t.listeners = append(t.listeners[:i], t.listeners[i+1:]...)
			close(l)
			return
		}
	}
}
