package logrusbugsnag

import (
	"errors"
	"runtime"

	"github.com/bugsnag/bugsnag-go"
	bugsnag_errors "github.com/bugsnag/bugsnag-go/errors"
	"github.com/sirupsen/logrus"
)

type Hook struct {
	mp *MetadataProvider
}

type MetadataProvider func(entry *logrus.Entry) bugsnag.MetaData

// ErrBugsnagUnconfigured is returned if NewBugsnagHook is called before
// bugsnag.Configure. Bugsnag must be configured before the hook.
var ErrBugsnagUnconfigured = errors.New("bugsnag must be configured before installing this logrus hook")

// ErrBugsnagSendFailed indicates that the hook failed to submit an error to
// bugsnag. The error was successfully generated, but `bugsnag.Notify()`
// failed.
type ErrBugsnagSendFailed struct {
	err error
}

func (e ErrBugsnagSendFailed) Error() string {
	return "failed to send error to Bugsnag: " + e.err.Error()
}

// NewBugsnagHook initializes a logrus hook which sends exceptions to an
// exception-tracking service compatible with the Bugsnag API. Before using
// this hook, you must call bugsnag.Configure(). The returned object should be
// registered with a log via `AddHook()`
//
// Entries that trigger an Error, Fatal or Panic should now include an "error"
// field to send to Bugsnag.
//
// Optionally pass a MetadataProvider to inject other custom metadata
func NewBugsnagHook(mp *MetadataProvider) (*Hook, error) {
	if bugsnag.Config.APIKey == "" {
		return nil, ErrBugsnagUnconfigured
	}
	hook := Hook{
		mp: mp,
	}
	return &hook, nil
}

// Fire forwards an error to Bugsnag. Given a logrus.Entry, it extracts the
// "error" field (or the Message if the error isn't present) and sends it off.
func (hook *Hook) Fire(entry *logrus.Entry) error {
	var notifyErr error
	err, ok := entry.Data["error"].(error)
	if ok {
		notifyErr = err
	} else {
		notifyErr = errors.New(entry.Message)
	}

	metadata := bugsnag.MetaData{}
	metadata["metadata"] = map[string]interface{}{
		"message": entry.Message,
	}
	for key, val := range entry.Data {
		if key != "error" {
			metadata["metadata"][key] = val
		}
	}

	if hook.mp != nil {
		for section, values := range (*hook.mp)(entry) {
			if _, ok := metadata[section]; !ok {
				metadata[section] = make(map[string]interface{})
			}
			for k, v := range values {
				metadata[section][k] = v
			}
		}
	}

	// if there's a panic on the stack (runtime.gopanic), assume we wanted
	// everything right before that.  Otherwise, assume we wanted everything 4+
	// frames up (before we got into logrus)
	depthOfPanic := findPanic()
	skipFrames := 5
	if depthOfPanic != 0 {
		skipFrames = depthOfPanic + 1
	}

	errWithStack := bugsnag_errors.New(notifyErr, skipFrames)
	bugsnagErr := bugsnag.Notify(errWithStack, metadata)
	if bugsnagErr != nil {
		return ErrBugsnagSendFailed{bugsnagErr}
	}

	return nil
}

const goPanic = "runtime.gopanic"

func findPanic() int {
	stack := make([]uintptr, 50)
	// skip two frames: runtime.Callers + findPanic
	nCallers := runtime.Callers(2, stack[:])
	frames := runtime.CallersFrames(stack[:nCallers])
	for i := 0; ; i++ {
		frame, more := frames.Next()
		if frame.Function == goPanic {
			return i
		}
		if !more {
			return 0
		}
	}
}

// Levels enumerates the log levels on which the error should be forwarded to
// bugsnag: everything at or above the "Error" level.
func (hook *Hook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}
}
