package logger

import (
	"context"
	"errors"
	"github.com/sirupsen/logrus"
)

var GlobalFields = logrus.Fields{}

type Logger func(context.Context, ...error) *logrus.Entry

func New(name string) Logger {
	return func(ctx context.Context, err ...error) *logrus.Entry {
		if len(err) == 1 && err[0] == nil {
			err = nil
		}
		return ContextLog(ctx, err, nil).WithField("component", name)
	}
}

type loggableError interface {
	error
	Loggable
}

func joinErrors(errs ...error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return errors.Join(errs...)
	}
}

func ContextLog(ctx context.Context, err []error, entry *logrus.Entry) *logrus.Entry {
	if entry == nil {
		entry = logrus.NewEntry(logrus.StandardLogger())
	}
	entry = entry.WithFields(GlobalFields)

	if ctx != nil {
		entry = entry.WithFields(GetLoggableValues(ctx))
		entry = entry.WithContext(ctx)
	}

	if err := joinErrors(err...); err != nil {
		entry = entry.WithError(err)

		// Check last, to allow LogFields to overwrite this package's behaviour.
		// Do not recurse in error causes, the error itself should merge its causes' fields if desired.
		var loggable loggableError
		if errors.As(err, &loggable) {
			entry = entry.WithFields(loggable.LogFields())
		}
	}

	return entry
}

// LogIfError makes it less verbose to defer a Close() call while
// handling an unlikely-but-possible error return by logging it. Example:
//
//	defer LogIfError(ctx, f.Close, log, "failed to close file")
func LogIfError(ctx context.Context, fn func() error, logger Logger, msg string) {
	if err := fn(); err != nil {
		logger(ctx, err).Error(msg)
	}
}
