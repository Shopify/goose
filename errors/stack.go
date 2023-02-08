package errors

import (
	stderrors "errors"
	"fmt"
	"runtime"
	"strings"

	pkgerrors "github.com/pkg/errors"
)

// findStackFromError returns the deepest stacktrace found. Support Courier errors and pkg/errors.
func findStackFromError(err error) []uintptr {
	if wrappedErr := stderrors.Unwrap(err); wrappedErr != nil {
		stack := findStackFromError(wrappedErr) // recursion
		if stack != nil {
			return stack // an inner error has a stacktrace, escape the recursion
		}
	}

	// starting from the inner error, look for stacktrace
	if stack := stackFromBaseError(err); stack != nil {
		return stack
	}
	return stackFromPkgError(err)
}

type pkgErrorWithStacktrace interface {
	StackTrace() pkgerrors.StackTrace
}

func stackFromPkgError(err error) []uintptr {
	var errWithStacktrace pkgErrorWithStacktrace
	if !stderrors.As(err, &errWithStacktrace) {
		return nil
	}

	stacktrace := errWithStacktrace.StackTrace()
	callers := make([]uintptr, len(stacktrace))
	for i, pc := range stacktrace {
		callers[i] = uintptr(pc) // de-alias from pkgerrors.Frame
	}
	return callers
}

func stackFromBaseError(err error) []uintptr {
	var errWithStacktrace *baseError
	if !stderrors.As(err, &errWithStacktrace) {
		return nil
	}

	return errWithStacktrace.stack
}

func formatStack(s []uintptr) string { // useful for debugging and writing tests.
	var builder strings.Builder

	frames := runtime.CallersFrames(s)
	for {
		frame, more := frames.Next()

		builder.WriteString(fmt.Sprintf("%s\n\t%s:%d", frame.Function, frame.File, frame.Line))
		builder.WriteString("\n")

		if !more {
			break
		}
	}

	return builder.String()
}
