package cond

import (
	"context"
	"sync"
	"time"

	"gopkg.in/tomb.v2"
)

type Cond struct {
	L     sync.Locker
	queue uint32

	// signal is the channel used by Signal to notify waiting threads.
	// Requires the lock to be held by Signal and NOT held by the waiters.
	signal chan struct{}

	// cancel is the channel used by the waiting threads to notify Signal that they are
	// no longer waiting for a signal, canceling the send.
	// This is necessary because there is a race condition where the waiter is no longer waiting,
	// but the queue count has not yet been decreased.
	// Requires the lock to be held by Signal and NOT held by the waiters.
	cancel chan struct{}
}

func NewCond(l sync.Locker) *Cond {
	return &Cond{
		L:      l,
		signal: make(chan struct{}),
		cancel: make(chan struct{}),
	}
}

// Wait unconditionally waits for a signal.
// Requires L to be locked.
func (c *Cond) Wait() {
	c.queue++
	c.L.Unlock()
	defer c.L.Lock()

	<-c.signal
}

// TimeoutWait waits for a signal for up to a duration.
// Requires L to be locked.
// Returns whether it received the signal.
func (c *Cond) TimeoutWait(d time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()

	return c.ContextWait(ctx)
}

// TombWait waits for a signal while the tomb is not dead.
// Requires L to be locked.
// Returns whether it received the signal.
func (c *Cond) TombWait(t *tomb.Tomb) bool {
	return c.waitChan(t.Dying())
}

// ContextWait waits for a signal while the context is not Done.
// Requires L to be locked.
// Returns whether it received the signal.
func (c *Cond) ContextWait(ctx context.Context) bool {
	return c.waitChan(ctx.Done())
}

// waitChan will wait for a signal or bail if ch is emitted.
// Requires L to be locked.
// Returns whether it received the signal.
func (c *Cond) waitChan(ch <-chan struct{}) bool {
	c.queue++
	c.L.Unlock()
	defer c.L.Lock()

	select {
	case <-ch:
		c.cancelWait()
		return false
	case <-c.signal:
		return true
	}
}

func (c *Cond) cancelWait() {
	select {
	case c.cancel <- struct{}{}:
		// We succeeded in sending the cancel signal to the thread that wanted to send us something.
		// That thread will decrease the queue count.
	default:
		// Nobody is listening to our cancel signal, decrease queue ourselves.
		c.L.Lock()
		c.queue--
		c.L.Unlock()
	}
}

// Signal will signal to a waiting thread, if any.
// Requires L to be locked.
func (c *Cond) Signal() {
	if c.queue == 0 {
		return
	}

	select {
	case c.signal <- struct{}{}:
	case <-c.cancel:
	}
	c.queue--
}

// Broadcast will signal to all waiting threads, if any.
// Requires L to be locked.
func (c *Cond) Broadcast() {
	for c.queue > 0 {
		select {
		case c.signal <- struct{}{}:
		case <-c.cancel:
		}
		c.queue--
	}
}
