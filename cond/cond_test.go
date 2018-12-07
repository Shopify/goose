package cond

import (
	"context"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"
)

const delay = 100 * time.Millisecond

func TestCond_Wait(t *testing.T) {
	c := NewCond(&sync.Mutex{})
	c.L.Lock()

	go func() {
		<-time.After(delay)
		c.L.Lock()
		c.Signal()
		c.L.Unlock()
	}()

	c.Wait()
	c.L.Unlock()
}

func TestCond_TimeoutWait(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := NewCond(&sync.Mutex{})
		c.L.Lock()

		go func() {
			<-time.After(delay)
			c.L.Lock()
			c.Signal()
			c.L.Unlock()
		}()

		assert.True(t, c.TimeoutWait(2*delay))
		c.L.Unlock()
	})

	t.Run("failure", func(t *testing.T) {
		c := NewCond(&sync.Mutex{})
		c.L.Lock()

		go func() {
			<-time.After(2 * delay)
			c.L.Lock()
			c.Signal()
			c.L.Unlock()
		}()

		assert.False(t, c.TimeoutWait(delay))
		c.L.Unlock()
	})
}

func TestCond_TombWait(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := NewCond(&sync.Mutex{})
		c.L.Lock()

		go func() {
			<-time.After(delay)
			c.L.Lock()
			c.Signal()
			c.L.Unlock()
		}()

		tb := &tomb.Tomb{}

		assert.True(t, c.TombWait(tb))
		c.L.Unlock()
	})

	t.Run("failure", func(t *testing.T) {
		c := NewCond(&sync.Mutex{})
		c.L.Lock()

		go func() {
			<-time.After(2 * delay)
			c.L.Lock()
			c.Signal()
			c.L.Unlock()
		}()

		tb := &tomb.Tomb{}
		go func() {
			<-time.After(delay)
			tb.Kill(nil)
		}()

		assert.False(t, c.TombWait(tb))
		c.L.Unlock()
	})
}

func TestCond_ContextWait(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := NewCond(&sync.Mutex{})
		c.L.Lock()

		go func() {
			<-time.After(delay)
			c.L.Lock()
			c.Signal()
			c.L.Unlock()
		}()

		ctx := context.Background()

		assert.True(t, c.ContextWait(ctx))
		c.L.Unlock()
	})

	t.Run("failure", func(t *testing.T) {
		c := NewCond(&sync.Mutex{})
		c.L.Lock()

		go func() {
			<-time.After(2 * delay)
			c.L.Lock()
			c.Signal()
			c.L.Unlock()
		}()

		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(delay))
		defer cancel()

		assert.False(t, c.ContextWait(ctx))
		c.L.Unlock()
	})
}

func TestCond_Signal(t *testing.T) {
	c := NewCond(&sync.Mutex{})
	c.L.Lock()

	go func() {
		<-time.After(delay)
		c.Wait()
		c.L.Unlock()
	}()

	// Same test as TestCond_Wait, but this time Signal() executes first
	c.L.Lock()
	c.Signal()
	c.L.Unlock()
}

func TestCond_Broadcast(t *testing.T) {
	c := NewCond(&sync.Mutex{})
	wg := sync.WaitGroup{}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			c.L.Lock()
			c.Wait()
			wg.Done()
			c.L.Unlock()
		}()
	}

	<-time.After(delay)
	c.L.Lock()
	c.Broadcast()
	c.L.Unlock()
	wg.Wait()
}

// Following tests are adapted for Go standard sync.Cond
// The only difference is that Signal and Broadcast must have the lock

func TestCondSignal(t *testing.T) {
	var m sync.Mutex
	c := NewCond(&m)
	n := 2
	running := make(chan bool, n)
	awake := make(chan bool, n)
	for i := 0; i < n; i++ {
		go func() {
			m.Lock()
			running <- true
			c.Wait()
			awake <- true
			m.Unlock()
		}()
	}
	for i := 0; i < n; i++ {
		<-running // Wait for everyone to run.
	}
	for n > 0 {
		select {
		case <-awake:
			t.Fatal("goroutine not asleep")
		default:
		}
		m.Lock()
		c.Signal()
		m.Unlock()
		<-awake // Will deadlock if no goroutine wakes up
		select {
		case <-awake:
			t.Fatal("too many goroutines awake")
		default:
		}
		n--
	}
	c.Signal()
}

func TestCondSignalGenerations(t *testing.T) {
	var m sync.Mutex
	c := NewCond(&m)
	n := 100
	running := make(chan bool, n)
	awake := make(chan int, n)
	for i := 0; i < n; i++ {
		go func(i int) {
			m.Lock()
			running <- true
			c.Wait()
			awake <- i
			m.Unlock()
		}(i)
		if i > 0 {
			a := <-awake
			if a != i-1 {
				t.Fatalf("wrong goroutine woke up: want %d, got %d", i-1, a)
			}
		}
		<-running
		m.Lock()
		c.Signal()
		m.Unlock()
	}
}

func TestCondBroadcast(t *testing.T) {
	var m sync.Mutex
	c := NewCond(&m)
	n := 200
	running := make(chan int, n)
	awake := make(chan int, n)
	exit := false
	for i := 0; i < n; i++ {
		go func(g int) {
			m.Lock()
			for !exit {
				running <- g
				c.Wait()
				awake <- g
			}
			m.Unlock()
		}(i)
	}
	for i := 0; i < n; i++ {
		for i := 0; i < n; i++ {
			<-running // Will deadlock unless n are running.
		}
		if i == n-1 {
			m.Lock()
			exit = true
			m.Unlock()
		}
		select {
		case <-awake:
			t.Fatal("goroutine not asleep")
		default:
		}
		m.Lock()
		c.Broadcast()
		m.Unlock()
		seen := make([]bool, n)
		for i := 0; i < n; i++ {
			g := <-awake
			if seen[g] {
				t.Fatal("goroutine woke up twice")
			}
			seen[g] = true
		}
	}
	select {
	case <-running:
		t.Fatal("goroutine did not exit")
	default:
	}
	c.Broadcast()
}

func TestRace(t *testing.T) {
	x := 0
	c := NewCond(&sync.Mutex{})
	done := make(chan bool)
	go func() {
		c.L.Lock()
		x = 1
		c.Wait()
		if x != 2 {
			t.Error("want 2")
		}
		x = 3
		c.Signal()
		c.L.Unlock()
		done <- true
	}()
	go func() {
		c.L.Lock()
		for {
			if x == 1 {
				x = 2
				c.Signal()
				break
			}
			c.L.Unlock()
			runtime.Gosched()
			c.L.Lock()
		}
		c.L.Unlock()
		done <- true
	}()
	go func() {
		c.L.Lock()
		for {
			if x == 2 {
				c.Wait()
				if x != 3 {
					t.Error("want 3")
				}
				break
			}
			if x == 3 {
				break
			}
			c.L.Unlock()
			runtime.Gosched()
			c.L.Lock()
		}
		c.L.Unlock()
		done <- true
	}()
	<-done
	<-done
	<-done
}

func TestCondSignalStealing(t *testing.T) {
	for iters := 0; iters < 1000; iters++ {
		var m sync.Mutex
		cond := NewCond(&m)

		// Start a waiter.
		ch := make(chan struct{})
		go func() {
			m.Lock()
			ch <- struct{}{}
			cond.Wait()
			m.Unlock()

			ch <- struct{}{}
		}()

		<-ch

		// We know that the waiter is in the cond.Wait() call because we
		// synchronized with it, then acquired/released the mutex it was
		// holding when we synchronized.
		//
		// Start two goroutines that will race: one will broadcast on
		// the cond var, the other will wait on it.
		//
		// The new waiter may or may not get notified, but the first one
		// has to be notified.
		done := false
		go func() {
			m.Lock()
			cond.Broadcast()
			m.Unlock()
		}()

		go func() {
			m.Lock()
			for !done {
				cond.Wait()
			}
			m.Unlock()
		}()

		// Check that the first waiter does get signaled.
		select {
		case <-ch:
		case <-time.After(2 * time.Second):
			t.Fatalf("First waiter didn't get broadcast.")
		}

		// Release the second waiter in case it didn't get the
		// broadcast.
		m.Lock()
		done = true
		cond.Broadcast()
		m.Unlock()
	}
}

func BenchmarkCond1(b *testing.B) {
	benchmarkCond(b, 1)
}

func BenchmarkCond2(b *testing.B) {
	benchmarkCond(b, 2)
}

func BenchmarkCond4(b *testing.B) {
	benchmarkCond(b, 4)
}

func BenchmarkCond8(b *testing.B) {
	benchmarkCond(b, 8)
}

func BenchmarkCond16(b *testing.B) {
	benchmarkCond(b, 16)
}

func BenchmarkCond32(b *testing.B) {
	benchmarkCond(b, 32)
}

func benchmarkCond(b *testing.B, waiters int) {
	c := NewCond(&sync.Mutex{})
	done := make(chan bool)
	id := 0

	for routine := 0; routine < waiters+1; routine++ {
		go func() {
			for i := 0; i < b.N; i++ {
				c.L.Lock()
				if id == -1 {
					c.L.Unlock()
					break
				}
				id++
				if id == waiters+1 {
					id = 0
					c.Broadcast()
				} else {
					c.Wait()
				}
				c.L.Unlock()
			}
			c.L.Lock()
			id = -1
			c.Broadcast()
			c.L.Unlock()
			done <- true
		}()
	}
	for routine := 0; routine < waiters+1; routine++ {
		<-done
	}
}
