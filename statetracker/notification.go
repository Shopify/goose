package statetracker

import "time"

// Notification is sent to listeners on all invocations of "Set"
type Notification struct {
	Old interface{}
	New interface{}

	// Time between Set(old) was called and Set(new).
	SinceOld time.Duration

	// Time between Set(new) was last called and this call.
	// Zero if Set(new) was never called before.
	// If new == old, SinceOld == SinceLastNew
	SinceLastNew time.Duration
}

func (n *Notification) IsInitial() bool {
	return n.SinceLastNew == 0
}
