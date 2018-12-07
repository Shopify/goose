package lockmap

import (
	"sync"
	"time"

	"gopkg.in/tomb.v2"
)

type Promise <-chan struct{}

type PromiseMap map[interface{}]Promise

type LockMap interface {
	// Wait returns the Promise for a key.
	// If the Promise is expired, it will be resolved, but this function will return nil
	Wait(key interface{}) Promise

	// WaitOrLock takes the lock, but only if none exists for that key.
	// If the Promise is expired, it will successfully replace it and the previous Promise will be resolved.
	WaitOrLock(key interface{}, ttl time.Duration) (promise Promise, gotLock bool)

	// Release unlocks a key and resolved the previous Promise, if any.
	Release(key interface{})

	// Allows the LockMap to be started and stopped externally
	Tomb() *tomb.Tomb
	Run() error
}

func New(sweepInterval time.Duration, tomb *tomb.Tomb) LockMap {
	return &lockMap{
		promises:      map[interface{}]*entry{},
		sweepInterval: sweepInterval,
		tomb:          tomb,
	}
}

type lockMap struct {
	// Use a lock and a regular map instead of a sync.Cond because some operations, like replace, are not available.
	l        sync.RWMutex
	promises map[interface{}]*entry

	sweepInterval time.Duration
	tomb          *tomb.Tomb
}

func (m *lockMap) Wait(key interface{}) Promise {
	now := time.Now().UnixNano()

	m.l.RLock()
	defer m.l.RUnlock()

	if prev, ok := m.promises[key]; ok {
		if prev.expiration > now {
			return prev.promise
		}
		prev.resolve()
	}

	return nil
}

func (m *lockMap) WaitOrLock(key interface{}, ttl time.Duration) (promise Promise, gotLock bool) {
	p := make(chan struct{})

	now := time.Now()
	expiration := now.Add(ttl).UnixNano()

	m.l.Lock()
	defer m.l.Unlock()

	if prev, ok := m.promises[key]; ok {
		if prev.expiration > now.UnixNano() {
			return prev.promise, false
		}

		// Current entry is expired
		prev.resolve()
	}

	m.promises[key] = &entry{
		expiration: expiration,
		promise:    p,
	}

	return p, true
}

func (m *lockMap) Release(key interface{}) {
	m.l.Lock()
	defer m.l.Unlock()

	if prev, ok := m.promises[key]; ok {
		prev.resolve()
	}

	delete(m.promises, key)
}

func (m *lockMap) Tomb() *tomb.Tomb {
	return m.tomb
}

func (m *lockMap) Run() error {
	ticker := time.NewTicker(m.sweepInterval)
	for {
		select {
		case <-m.tomb.Dying():
			ticker.Stop()
			m.shutdown()
			return m.tomb.Err()
		case <-ticker.C:
			m.sweep()
		}
	}
}

func (m *lockMap) shutdown() {
	m.l.Lock()
	defer m.l.Unlock()

	for _, entry := range m.promises {
		entry.resolve()
	}
}

func (m *lockMap) sweep() {
	now := time.Now().UnixNano()

	m.l.Lock()
	defer m.l.Unlock()

	for key, entry := range m.promises {
		if entry.expiration <= now {
			entry.resolve()
			delete(m.promises, key)
		}
	}
}
