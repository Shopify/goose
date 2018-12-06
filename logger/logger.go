package logger

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type causer interface {
	Cause() error
}

var GlobalFields = logrus.Fields{}

type Logger func(Valuer, error) *logrus.Entry

func New(name string) Logger {
	return func(ctx Valuer, err error) *logrus.Entry {
		return ContextLog(ctx, err, nil).WithField("component", name)
	}
}

func ContextLog(ctx Valuer, err error, entry *logrus.Entry) *logrus.Entry {
	if entry == nil {
		entry = logrus.NewEntry(logrus.StandardLogger())
	}
	entry = entry.WithFields(GlobalFields)

	if ctx != nil {
		entry = entry.WithFields(getLoggableValues(ctx))
	}

	if err != nil {
		entry = entry.WithField("error", err)

		if _, ok := err.(causer); ok {
			entry = entry.WithField("cause", errors.Cause(err))
		}
	}

	return entry
}

// LogIfError makes it less verbose to defer a Close() call while
// handling an unlikely-but-possible error return by logging it. Example:
//
//   defer LogIfError(f.Close, log, "failed to close file")
func LogIfError(fn func() error, logger Logger, msg string) {
	if err := fn(); err != nil {
		logger(nil, err).Error(msg)
	}
}
