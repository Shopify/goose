package errors

import (
	"context"
	"fmt"
	"runtime"

	bugsnag_errors "github.com/bugsnag/bugsnag-go/v2/errors"
	"github.com/sirupsen/logrus"
)

// Fields can be attached to errors like this:
// errors.Wrap(err, "invalid value", errors.Fields{"value": value})
type Fields map[string]interface{}

type ErrorWithDetails interface {
	ErrorDetails() map[string]interface{}
}

type baseError struct {
	err     error
	message string
	fields  Fields
	stack   []uintptr
}

func (e *baseError) Error() string {
	return e.message
}

func (e *baseError) Unwrap() error {
	return e.err
}

func (e *baseError) Callers() []uintptr {
	return findStackFromError(e) // lazy stack lookup in wrapped errors
}

func (e *baseError) LogFields() logrus.Fields {
	return logrus.Fields(e.fields)
}

var _ bugsnag_errors.ErrorWithCallers = &baseError{}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
func Errorf(format string, args ...interface{}) error {
	return &baseError{
		message: fmt.Sprintf(format, args...),
		stack:   captureStack(),
	}
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
func Wrap(err error, message string, fields ...Fields) error {
	if err == nil {
		return nil
	}

	return &baseError{
		err:     err,
		message: fmt.Sprintf("%s: %s", message, err.Error()),
		fields:  mergeFieldsCtx(nil, err, fields...),
		stack:   captureStack(),
	}
}

// WrapCtx returns an error annotating err with a stack trace and log fields.
// The log fields are captured from context.Context and arguments.
// If err is nil, WrapCtx returns nil.
func WrapCtx(ctx context.Context, err error, message string, fields ...Fields) error {
	if err == nil {
		return nil
	}

	return &baseError{
		err:     err,
		message: message + ": " + err.Error(),
		fields:  mergeFieldsCtx(ctx, err, fields...),
		stack:   captureStack(),
	}
}

// With returns an error annotating err with a stack trace and log fields.
// If err is nil, With returns nil.
func With(err error, fields ...Fields) error {
	if err == nil {
		return nil
	}

	return &baseError{
		err:     err,
		message: err.Error(),
		fields:  mergeFieldsCtx(nil, err, fields...),
		stack:   captureStack(),
	}
}

// WithCtx returns an error annotating err with a stack trace and log fields.
// The log fields are captured from context.Context and arguments.
// If err is nil, WithCtx returns nil.
func WithCtx(ctx context.Context, err error, fields ...Fields) error {
	if err == nil {
		return nil
	}

	return &baseError{
		err:     err,
		message: err.Error(),
		fields:  mergeFieldsCtx(ctx, err, fields...),
		stack:   captureStack(),
	}
}

func captureStack() []uintptr {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // Wrap/WrapCtx -> findStack -> runtime.Callers
	return pcs[0:n]
}
