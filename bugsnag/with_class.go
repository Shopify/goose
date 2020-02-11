package bugsnag

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// An interface errors can implement to better control how they are
// grouped in Bugsnag.
// https://docs.bugsnag.com/product/error-grouping/#top-in-project-stackframe
type errorClasser interface {
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
	return w.cause
}

func (w *withClass) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, w.class)
			_, _ = io.WriteString(s, ": ")
			_, _ = fmt.Fprintf(s, "%+v\n", w.Cause())
			return
		}
		fallthrough
	case 's', 'q':
		_, _ = io.WriteString(s, w.Error())
	}
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

// Wrapf acts like errors.Wrapf except it also sets the
// error class to be equal to the message
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf(format, args...)
	return WithErrorClass(errors.Wrap(err, message), message)
}

func extractErrorClass(err error) string {
	type causer interface {
		Cause() error
	}

	for err != nil {
		if classer, ok := err.(errorClasser); ok {
			return classer.ErrorClass()
		}

		if causer, ok := err.(causer); ok {
			err = causer.Cause()
		} else {
			return err.Error()
		}
	}

	return ""
}
