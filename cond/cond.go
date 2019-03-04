package cond

import (
	"context"
	"sync"
	"time"

	"gopkg.in/tomb.v2"
)

type Cond struct {
	L      sync.Locker
	queue  uint32
	signal chan struct{}
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
	c.queue++
	c.L.Unlock()
	defer c.L.Lock()

	t := time.NewTimer(d)
	defer t.Stop()

	select {
	case <-t.C:
		c.cancelWait()
		return false
	case <-c.signal:
		return true
	}
}

// TombWait waits for a signal while the tomb is not dead.
// Requires L to be locked.
// Returns whether it received the signal.
func (c *Cond) TombWait(t *tomb.Tomb) bool {
	c.queue++
	c.L.Unlock()
	defer c.L.Lock()

	select {
	case <-t.Dying():
		c.cancelWait()
		return false
	case <-c.signal:
		return true
	}
}

// ContextWait waits for a signal while the context is not Done.
// Requires L to be locked.
// Returns whether it received the signal.
func (c *Cond) ContextWait(ctx context.Context) bool {
	c.queue++
	c.L.Unlock()
	defer c.L.Lock()

	select {
	case <-ctx.Done():
		c.cancelWait()
		return false
	case <-c.signal:
		return true
	}
}

func (c *Cond) cancelWait() {
	gotLock := make(chan struct{})
	cancelled := make(chan bool)

	go func() {
		c.L.Lock()
		close(gotLock)
		if <-cancelled {
			c.L.Unlock()
			return
		}
		c.queue--
		c.L.Unlock()
	}()

	select {
	case c.cancel <- struct{}{}:
		cancelled <- true
		// We succeeded in sending the cancel signal to the thread that wanted to send us something.
		// That thread will decrease the queue count.
	case <-gotLock:
		cancelled <- false
		// Nobody was listening to our cancel signal, the goroutine will decrease the queue count.
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
