/*
Package bugsnag provides hooks on the bugsnag library.

# Usage

The client should only need to use the static/package level APIs defined in bugsnag.go

```go

	import "github.com/Shopify/goose/v2/bugsnag"
	import bugsnaggo "github.com/bugsnag/bugsnag-go/v2"

	func main(){
		bugsnag.AutoConfigure("apiKey", "SHA", "env", []string{"github.com/Shopify/yourClientRepo"})
		defer bugsnaggo.AutoNotify()

		i, err := somethingThatCanError()
		bugsnaggo.Notify(err, rawData...)
	}

```

# Features

Smart handling of errors passed to notify populating the "error" tab in bugsnag. If the error is wrapped using pkg/errors then its stacktrace
and context messages are extracted correctly.

You can use bugsnag.WithErrorClass(err, class) or implement the ErrorClassProvider interface to better control how the error is grouped in Bugsnag.

Better support for logrus.Fields, *logrus.Entry, and *http.Request

HTTPRequestFormHook is available to extract the HTTP POST request body into the request tab in bugsnag, but it is not available by default.

Project packages are handled correctly which helps with proper grouping of bugs in bugsnag.

# A panic handler is available, but not enabled by default, since it has the tendency to not play nice with many go invocation methods

Create custom bugsnag tabs easily by implementing the TabProvider interface or Tab objects passed in as rawData. e.g.:

	type tabProvider struct {
		val string
	}

	func (ct *tabProvider) CreateBugsnagTab() Tab {
		return Tab{
			Label: "Custom Tab",
			Rows:  Rows{"Key": ct.val},
		}
	}

	bugsnag.Notify(err, &tabProvider{"value"})
*/
package bugsnag
