package logger

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
)

type loggerContextKeyType struct{}

var (
	loggerContextKey = loggerContextKeyType{}
)

func WithLogger(ctx context.Context, logger FieldLogger) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

func FromContext(ctx context.Context) FieldLogger {
	if ctx == nil {
		ctx = context.Background()
	}

	logger, _ := ctx.Value(loggerContextKey).(FieldLogger)
	if logger == nil {
		logger = DefaultLogger()
	}
	return logger.WithContext(ctx)
}

type FieldLogger interface {
	logrus.FieldLogger
	// WithContext Unfortunately, logrus' signature returns a *logrus.Entry, so we can't return a FieldLogger
	WithContext(ctx context.Context) *logrus.Entry
}

type Logger func(context.Context, ...error) *logrus.Entry

func New(name string) Logger {
	return func(ctx context.Context, err ...error) *logrus.Entry {
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
		entry = FromContext(ctx).WithFields(logrus.Fields{}) // Call WithFields to get an Entry
	}

	if err := joinErrors(err...); err != nil {
		entry = entry.WithError(err)
	}

	return entry
}

// LogIfError makes it less verbose to defer a Close() call while
// handling an unlikely-but-possible error return by logging it. Example:
//
//	defer LogIfError(ctx, f.Close, log, "failed to close file")
func LogIfError(ctx context.Context, fn func() error, logger Logger, msg string) {
	if err := fn(); err != nil {
		if logger != nil {
			logger(ctx, err).Error(msg)
		} else {
			FromContext(ctx).WithError(err).Error(msg)
		}
	}
}
