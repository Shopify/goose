package genmain

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Shopify/goose/logger"
	"github.com/Shopify/goose/metrics"
	"github.com/Shopify/goose/safely"
	"github.com/Shopify/goose/statsd"
)

var log = logger.New("genmain")

const defaultShutdownDeadline = 3 * time.Second

// Component is used to represent various "components". At a high level, main()
// essentially cobbles together a few components whose lifecycles are managed
// by Tombs. `Component` allows us to treat them as black boxes.
type Component safely.Runnable

var (
	// ErrAlreadyRunning is returned when `RunAndWait` is called after already
	// being called.
	ErrCanOnlyRunOnce = errors.New("can only run once")

	// ErrShutdownRequested can be used as a reason for `Kill` that indicates no
	// error has occurred, just that the components should gracfeully exit.
	ErrShutdownRequested = errors.New("shutdown requested")
)

// Main represents a collection of components whose lifecycles are tied together.
type Main struct {
	components       []Component
	shutdownDeadline time.Duration

	l   sync.Mutex
	ran bool
}

// New creates a new `Main`
func New(components ...Component) Main {
	return Main{
		components:       components,
		shutdownDeadline: defaultShutdownDeadline,
	}
}

// SignalError is an `error` returned when `Main` exits due to a signal.
type SignalError struct {
	signal os.Signal
}

func (s *SignalError) Error() string {
	return fmt.Sprintf("received signal: %v", s.signal)
}

// Kill will terminate all running components with a given reason
func (m *Main) Kill(reason error) {
	// Acquire the lock to ensure the first call's `err` is the one that all components
	// receive, and not some mix if this function was called concurrently.
	m.l.Lock()
	defer m.l.Unlock()

	log(nil, reason).Info("shutting down")
	for _, c := range m.components {
		c.Tomb().Kill(reason)
	}
}

func (m *Main) SetShutdownDeadline(d time.Duration) {
	m.shutdownDeadline = d
}

// RunAndWait starts all components in this `Main`.
//
// `RunAndWait` will also listen to SIGINT and SIGTERM to do graceful shutdowns of all
// components it manages. It should only be called once, and returns `ErrCanOnlyRunOnce`
// if called more than once.
func (m *Main) RunAndWait() error {
	ctx := context.Background()
	defer metrics.GenMainRun.StartTimer(ctx).Finish()

	m.l.Lock()
	if m.ran {
		m.l.Unlock()
		return ErrCanOnlyRunOnce
	}

	m.ran = true
	m.l.Unlock()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	shutdown := make(chan error, len(m.components)+1)
	safely.Go(func() {
		sig := <-sigs
		go func() {
			<-sigs
			log(nil, nil).Fatal("received signal again: terminating immediately")
		}()

		shutdown <- &SignalError{sig}
	})

	for _, c := range m.components {
		safely.Run(c)

		go func(comp Component) {
			<-comp.Tomb().Dying()
			log(nil, comp.Tomb().Err()).
				WithField("mainComponent", componentName(comp)).
				Info("process exited")
			shutdown <- ErrShutdownRequested
		}(c)
	}

	reason := <-shutdown

	shutdownStart := time.Now()
	deadline := time.After(m.shutdownDeadline)

	m.Kill(reason)

	for _, comp := range m.components {
		ctx = statsd.WithTagLogFields(ctx, logrus.Fields{
			"deadline":      m.shutdownDeadline,
			"mainComponent": componentName(comp),
		})
		select {
		case <-comp.Tomb().Dead():
			metrics.GenMainShutdown.Duration(ctx, time.Since(shutdownStart), statsd.Tags{"success": true})
			log(ctx, comp.Tomb().Err()).Debug("component shutdown notification")
		case <-deadline:
			metrics.GenMainShutdown.Duration(ctx, time.Since(shutdownStart), statsd.Tags{"success": false})
			log(ctx, comp.Tomb().Err()).Error("component took too long to shut down")
		}
	}

	log(nil, nil).Debug("final shutdown message")

	return reason
}

func componentName(comp Component) string {
	compType := fmt.Sprintf("%T", comp)
	return strings.TrimPrefix(compType, "*")
}
