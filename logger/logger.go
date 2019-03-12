package logger

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type causer interface {
	Cause() error
}

var GlobalFields = logrus.Fields{}

type Logger func(Valuer, ...error) *logrus.Entry

func New(name string) Logger {
	return func(ctx Valuer, err ...error) *logrus.Entry {
		if len(err) == 1 && err[0] == nil {
			err = nil
		}
		return ContextLog(ctx, err, nil).WithField("component", name)
	}
}

func ContextLog(ctx Valuer, err []error, entry *logrus.Entry) *logrus.Entry {
	if entry == nil {
		entry = logrus.NewEntry(logrus.StandardLogger())
	}
	entry = entry.WithFields(GlobalFields)

	if ctx != nil {
		entry = entry.WithFields(getLoggableValues(ctx))
		if ctx, ok := ctx.(context.Context); ok {
			entry = entry.WithContext(ctx)
		}
	}

	if len(err) != 0 {
		err0 := err[0]
		entry = entry.WithField("error", err0)

		if _, ok := err0.(causer); ok {
			entry = entry.WithField("cause", errors.Cause(err0))
		}

		for i, errX := range err[1:] {
			entry = entry.WithField(fmt.Sprintf("error%d", i+1), errX)

			if _, ok := errX.(causer); ok {
				entry = entry.WithField(fmt.Sprintf("cause%d", i+1), errors.Cause(errX))
			}
		}
	}

	return entry
}

// LogIfError makes it less verbose to defer a Close() call while
// handling an unlikely-but-possible error return by logging it. Example:
//
//   defer LogIfError(ctx, f.Close, log, "failed to close file")
func LogIfError(ctx context.Context, fn func() error, logger Logger, msg string) {
	if err := fn(); err != nil {
		logger(ctx, err).Error(msg)
	}
}
