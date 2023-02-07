// Package bugsnag extends the functionalities of the APIs in the bugsnag-go library.
package bugsnag

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"
	bugsnaggoErr "github.com/bugsnag/bugsnag-go/v2/errors"
	"github.com/bugsnag/panicwrap"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
)

// Rows is a helper type to make it easier to build bugsnag tab contents,
// with simple Go literals as in `Tab{"Stuff",Rows{"a":1,"b":2,...}}`.
type Rows map[string]interface{}

// Tab can be used to attach an additional info tab to bugsnag error reports.
// Just pass it to the Notify function.
type Tab struct {
	Label string
	Rows  Rows
}

// notifier by default is linked to bugsnaggo.Notify. We swap it out for testing.
type notifier func(err error, rawData ...interface{}) (e error)

// Bugsnagger uses the default Notifier and Config object that bugsnag-go (3rd party lib) provides.
// It formats the rawData to the general MetaData format for the bugsnag-go API.
// Bugsnagger also contains the implementation for the external APIs that are defined in bugsnag.go.
type bugsnagger struct {
	config   *bugsnaggo.Configuration
	notifier notifier
}

func newBugsnagger(config *bugsnaggo.Configuration, notifier notifier) *bugsnagger {
	return &bugsnagger{config: config, notifier: notifier}
}

// Optional metadata can be passed in as well, however only certain types of those will have an effect:
// * string will be used to set the BS context (usually url path or goroutine identifier)
// * http.Request creates Request tab in BS and tries to set context
// * log.Entry/log.Fields creates Details tab reflecting the entry fields
// * Tab allows creating arbitrary tab
func (snagger *bugsnagger) buildData(err error, rawData ...interface{}) ([]interface{}, error) {
	var (
		user    bugsnaggo.User
		req     *http.Request
		context string
		cause   error
	)

	bugsnagData := []interface{}{bugsnaggo.SeverityError}
	md := make(bugsnaggo.MetaData)

	if err != nil {
		bugsnagData = append(bugsnagData, bugsnaggo.ErrorClass{Name: extractErrorClass(err)})
		context, cause = snagger.formatError(err)
		// Set the Error tab with details/stacktrace
		md["Error"] = Rows{"details": fmt.Sprintf("%+v", err)}
	}

	// Extract more context/metadata from the rawData
	for _, v := range rawData {
		switch v := v.(type) {
		case string:
			context = v
		case *http.Request:
			req = v
			bugsnagData = append(bugsnagData, req)
		case *log.Entry:
			userID, message, mdData := snagger.processLogEntry(v)
			snagger.updateUserID(&user, userID)
			if context == "" {
				context = message
			}
			md["Log"] = mdData
		case log.Fields:
			userID := snagger.extractUserID(v)
			snagger.updateUserID(&user, userID)
			md["Log"] = v
		case TabWriter:
			tab := v.CreateBugsnagTab()
			md[tab.Label] = tab.Rows
		case Tab:
			md[v.Label] = v.Rows
		default:
			bugsnagData = append(bugsnagData, v)
		}
	}

	if len(md) > 0 {
		bugsnagData = append(bugsnagData, md)
	}
	if user.Id != "" {
		bugsnagData = append(bugsnagData, user)
	}
	if context == "" && req != nil && req.URL != nil {
		context = req.URL.Path
	}
	bugsnagData = append(bugsnagData, bugsnaggo.Context{String: context})

	return bugsnagData, cause
}

func (snagger *bugsnagger) formatError(err error) (string, error) {
	var context string
	cause := errors.Cause(err)
	if uerr, ok := cause.(*url.Error); ok {
		cause = uerr.Err
	} else if serr, ok := cause.(*http2.StreamError); ok {
		if serr.Cause != nil {
			cause = serr.Cause
		} else {
			cause = fmt.Errorf("stream error: %s", serr.Code)
		}
	}

	// Try to get the wrapped message to set the bugsnag context
	errString := fmt.Sprintf("%s", err)
	// Note that error.Wrap always appends a space after the :
	i := strings.LastIndex(errString, ": ")
	if i != -1 {
		context = errString[0:i]
	}
	return context, cause
}

func (snagger *bugsnagger) processLogEntry(entry *log.Entry) (string, string, map[string]interface{}) {
	fields := entry.Data
	return snagger.extractUserID(fields), entry.Message, fields
}

func (snagger *bugsnagger) extractUserID(fields log.Fields) string {
	var userID string
	if s := fields["shopID"]; s != nil {
		userID = fmt.Sprint(s)
	}
	return userID
}

func (snagger *bugsnagger) updateUserID(user *bugsnaggo.User, userID string) {
	if len(userID) > 0 {
		user.Id = userID
	}
}

// Notify annotates bugsnag entries with metadata and then sends it to bugsnag
func (snagger *bugsnagger) Notify(err error, rawData ...interface{}) {
	if err == nil || snagger.config.APIKey == "" {
		return
	}

	bugsnagData, cause := snagger.buildData(err, rawData...)
	if err := snagger.notifier(cause, bugsnagData...); err != nil {
		log.Warnf("bugsnag notifier error: %s\n", err)
	}
}

// AutoNotify when deferred, records a panic to Bugsnag and exits the program
func (snagger *bugsnagger) AutoNotify(rawData ...interface{}) {
	if r := recover(); r != nil {
		snagger.Notify(recoverError(r), rawData...)
		log.Error(r)
		debug.PrintStack()
		os.Exit(1)
	}
}

// AutoRecover when deferred, records a panic to Bugsnag and recovers.
func (snagger *bugsnagger) AutoRecover(rawData ...interface{}) {
	if r := recover(); r != nil {
		snagger.Notify(recoverError(r), rawData...)
		log.Error(r)
	}
}

func recoverError(e interface{}) error {
	var err error
	switch e := e.(type) {
	case error:
		err = e
	default:
		err = errors.Errorf("%v (%T)", e, e)
	}
	return err
}

// Augment the bugsnag http middleware to add POST body to the request tab
func httpRequestMiddleware(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
	for _, datum := range event.RawData {
		if request, ok := datum.(*http.Request); ok {
			if err := request.ParseForm(); err != nil {
				return err
			}
			event.MetaData.Update(bugsnaggo.MetaData{
				"request": {
					"form": request.Form,
				},
			})
		}
	}
	return nil
}

func (snagger *bugsnagger) Setup(apiKey string, commit string, env string, packages []string) {
	// Add the bugsnag package and it's folder location on disk to bugsnag's ProjectPackages.
	// This will ensure that Notify calls from bugsnagger.go will always share the same file name
	// and will retain grouping across Shopify/goose dependency upgrades.
	packages = append(packages, "main*", "github.com/Shopify/goose/v2/bugsnag")
	if _, file, _, ok := runtime.Caller(0); ok {
		gooseMod := strings.TrimSuffix(file, "bugsnag/bugsnagger.go")
		packages = append(packages, gooseMod+"*")
	}

	bugsnaggo.OnBeforeNotify(httpRequestMiddleware)
	bugsnaggo.Configure(bugsnaggo.Configuration{
		APIKey:          apiKey,
		AppVersion:      commit,
		ProjectPackages: packages,
		ReleaseStage:    env,
		Synchronous:     true,
		PanicHandler:    panicHandler,
	})
}

// panicHandler uses panicwrap.BasicWrap instead of panicwrap.BasicMonitor to catch panics and notify.
// The difference is that the parent process is now the one monitoring and so there is no race when the application
// is containerized.
func panicHandler() {
	defer AutoNotify()

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
		Notify(toNotify, state, bugsnaggo.Configuration{Synchronous: true})
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
