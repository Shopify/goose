package bugsnag

import (
	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"
	"github.com/pkg/errors"

	"github.com/Shopify/goose/v2/metrics"
)

// ErrorClassProvider can be implemented to better control how the error is
// grouped in Bugsnag.
// https://docs.bugsnag.com/product/error-grouping/#top-in-project-stackframe
type ErrorClassProvider interface {
	ErrorClass() string
}

type withClass struct {
	cause error
	class string
}

func (w *withClass) Error() string {
	return w.cause.Error()
}

func (w *withClass) Cause() error {
	return w.Unwrap()
}

func (w *withClass) Unwrap() error {
	return w.cause
}

func (w *withClass) ErrorClass() string {
	return w.class
}

// WithErrorClass wraps the error with an error class used
// to control the grouping in Bugsnag
func WithErrorClass(err error, class string) error {
	if err == nil {
		return nil
	}
	return &withClass{
		cause: err,
		class: class,
	}
}

func ErrorClassHook(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
	var errClasser ErrorClassProvider
	if errors.As(event.Error.Err, &errClasser) {
		errorClass := errClasser.ErrorClass()
		event.ErrorClass = errorClass
		event.Ctx = metrics.WithTag(event.Ctx, "error_class", errorClass)
	}
	return nil
}
