package bugsnag

import (
	"os"

	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"
	bugsnaggoErr "github.com/bugsnag/bugsnag-go/v2/errors"
	"github.com/bugsnag/panicwrap"
	log "github.com/sirupsen/logrus"
)

// BasicWrapPanicHandler uses panicwrap.BasicWrap instead of panicwrap.BasicMonitor to catch panics and notify.
// The difference is that the parent process is now the one monitoring and so there is no race when the application
// is containerized.
func BasicWrapPanicHandler() {
	defer bugsnaggo.AutoNotify()

	exitStatus, err := panicwrap.BasicWrap(func(output string) {
		toNotify, err := bugsnaggoErr.ParsePanic(output)

		if err != nil {
			log.Errorf("bugsnag.handleUncaughtPanic: %v", err)
		}
		state := bugsnaggo.HandledState{
			SeverityReason:   bugsnaggo.SeverityReasonUnhandledPanic,
			OriginalSeverity: bugsnaggo.SeverityError,
			Unhandled:        true,
			Framework:        "",
		}
		_ = bugsnaggo.Notify(toNotify, state, bugsnaggo.Configuration{Synchronous: true})
	})
	if err != nil {
		// Something went wrong setting up the panic wrapper. Unlikely,
		// but possible.
		panic(err)
	}

	// If exitStatus >= 0, then we're the parent process and the panicwrap
	// re-executed ourselves and completed. Just exit with the proper status.
	if exitStatus >= 0 {
		os.Exit(exitStatus) // nolint:gocritic
	}

	// Otherwise, exitStatus < 0 means we're the child. Continue executing as
	// normal...
}
