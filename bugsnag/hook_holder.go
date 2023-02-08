package bugsnag

import (
	"errors"
	"sync"
	"sync/atomic"

	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"
)

var (
	ErrCaptured = errors.New("bugsnag report captured")
	GlobalHook  HookHolder
)

// HookHolder is designed to be an overridable hook.
// It can be used in testing to override the GlobalHook to swallow all reporting
type HookHolder struct {
	l        sync.RWMutex
	skipLock atomic.Bool
	hook     Hook
}

func (i *HookHolder) CaptureEvents(fn func()) (events []*bugsnaggo.Event) {
	i.WithHook(func(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
		events = append(events, event)
		return ErrCaptured
	}, fn)

	return events
}

func (i *HookHolder) WithHook(hook Hook, fn func()) {
	i.l.Lock()
	i.skipLock.Store(true)
	defer func() {
		i.skipLock.Store(false)
		i.l.Unlock()
	}()

	prev := i.hook
	defer func() {
		i.hook = prev
	}()

	i.hook = hook
	fn()
}

func (i *HookHolder) SetHook(hook Hook) {
	i.l.Lock()
	defer i.l.Unlock()

	i.hook = hook
}

func (i *HookHolder) Hook(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
	if !i.skipLock.Load() {
		i.l.RLock()
		defer i.l.RUnlock()
	}

	if i.hook == nil {
		return nil
	}

	return i.hook(event, config)
}
