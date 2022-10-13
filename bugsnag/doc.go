/*
Package bugsnag provides a simple wrapper over the bugsnag/bugsnag-go lib that maintains
the same Notify(err, rawData...) API but with improved setup and error decoration.

# Usage

The client should only need to use the static/package level APIs defined in bugsnag.go

	func main(){
		bugsnag.Setup("apiKey", "SHA", "env", []string{"github.com/Shopify/yourClientRepo"})
		defer bugsnag.AutoNotify()

		go func(){
			defer bugsnag.AutoRecover()
			panic("err")
		}

		i, err := somethingThatCanError()
		bugsnag.Notify(err, rawData...)
	}

# Features

Smart handling of errors passed to notify populating the "Error" tab in bugsnag. If the error is wrapped using pkg/errors then its stacktrace
and context messages are extracted correctly.

The bugsnag.ErrorClass is set intelligently dealing with go.mod versions. You can also implement the errorClasser interface
to override this using the bugsnag.WithErrorClass(err, class) and bugsnag.Wrapf(err, format, args...) helper methods.

Better support for logrus.Fields, *logrus.Entry, *http.Request, *url.Error and *http2.StreamError objects.

HTTP POST request body is extracted into the the Request tab in bugsnag.

Project packages are handled correctly which helps with proper grouping of bugs in bugsnag.

The default panic handler is fixed so that panics are not dropped in containerized environments.

Create custom bugsnag tabs easily by implementing the TabWriter interface or Tab objects passed in as rawData. e.g.:

	type customTab struct {
		val string
	}

	func (ct *customTab) CreateBugsnagTab() Tab {
		return Tab{
			Label: "Custom Tab",
			Rows:  Rows{"Key": ct.val},
		}
	}

	bugsnag.Notify(err, &customTab{"value"})
*/
package bugsnag
