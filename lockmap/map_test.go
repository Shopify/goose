package lockmap

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/tomb.v2"
)

func ExampleNew() {
	m := New(1*time.Minute, &tomb.Tomb{})
	m.Tomb().Go(m.Run)

	output := ""

	done := make(chan struct{})
	waitOrWork := func() {
		promise, gotLock := m.WaitOrLock("foo", 10*time.Second)
		if gotLock {
			// We just grabbed a new lock, do work with it.
			<-time.After(100 * time.Millisecond)
			output = "done"

			// Once we're done, remove it, which will close the channel on the promise.
			m.Release("foo")
		} else {
			<-promise
			fmt.Println(output)
			close(done)
		}
	}

	// Either one could be trigger first
	go waitOrWork()
	go waitOrWork()

	<-done
	// Output:
	// done
}

const sweepInterval = 1 * time.Second
const ttl = sweepInterval / 3

func newMap(interval time.Duration) LockMap {
	m := New(interval, &tomb.Tomb{})
	m.Tomb().Go(m.Run)
	return m
}

func assertResolved(promise Promise) {
	select {
	case <-promise:
	default:
		panic("promise should have been resolved")
	}
}

func assertNotResolved(promise Promise) {
	select {
	case <-promise:
		panic("promise should not have been resolved")
	default:
	}
}

func Test_lockMap_Release(t *testing.T) {
	m := newMap(sweepInterval)
	promise, gotLock := m.WaitOrLock(1, ttl)
	assert.True(t, gotLock)

	stored := m.Wait(1)
	assert.Equal(t, promise, stored)

	m.Release(1)

	stored = m.Wait(1)
	assert.Nil(t, stored)

	assertResolved(promise)
}

func Test_lockMap_Wait(t *testing.T) {
	m := newMap(sweepInterval)
	promise, gotLock := m.WaitOrLock(1, ttl)
	assert.True(t, gotLock)

	stored := m.Wait(1)
	assert.Equal(t, promise, stored)

	<-time.After(ttl)

	stored = m.Wait(1)
	assert.Nil(t, stored)

	assertResolved(promise)
}

func Test_lockMap_WaitOrLock(t *testing.T) {
	m := newMap(sweepInterval)

	promise, gotLock := m.WaitOrLock(1, ttl)
	assert.True(t, gotLock)

	stored, gotLock := m.WaitOrLock(1, ttl)
	assert.False(t, gotLock)
	assert.Equal(t, promise, stored)

	assertNotResolved(promise)
}

func Test_lockMap_sweep(t *testing.T) {
	sweep := ttl * 3 / 2
	m := newMap(sweep)

	promise, _ := m.WaitOrLock(1, ttl)
	promise2, _ := m.WaitOrLock(2, ttl*3)

	<-time.After(sweep + 100*time.Millisecond)

	assertResolved(promise)
	assertNotResolved(promise2)

	stored := m.Wait(1)
	assert.Nil(t, stored)

	stored = m.Wait(2)
	assert.Equal(t, promise2, stored)
}
