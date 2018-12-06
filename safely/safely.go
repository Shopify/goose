package safely

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/tomb.v2"
)

// Go executes a supplied function and recovers from any panics that happen,
// logging those panics with a severity sufficient to push them to whatever
// error recording systems are registered as logrus hooks (i.e.  bugsnag).
// It's meant to be called like: go safely.Go(...)
func Go(f func()) {
	go func() {
		defer Recover()
		f()
	}()
}

// ErrPanicked is passed to bugsnag when we instrument a panic.
type ErrPanicked struct {
	val interface{}
}

func (e *ErrPanicked) Error() string {
	return fmt.Sprintf("panic(%#v)", e.val)
}

// Recover recovers a panic, if any, and submits it to bugsnag through logrus
// (assuming that handler is registered) before terminating the program for
// real.
func Recover() {
	if r := recover(); r != nil {
		log.WithField("error", &ErrPanicked{r}).Panicf("intercepted panic via safely.Recover")
	}
}

// TombGo is the compose of Go and tomb.Go. Unlike Go though, the supplied
// function must return an error, because that's what tomb expects.
func TombGo(t *tomb.Tomb, f func() error) {
	t.Go(func() error {
		defer Recover()
		return f()
	})
}

type Runnable interface {
	// Tomb returns this Runnable's tomb used for lifecycle management.
	Tomb() *tomb.Tomb

	// Run begins the Runnable's main run loop.
	Run() error
}

func Run(r Runnable) {
	TombGo(r.Tomb(), r.Run)
}
