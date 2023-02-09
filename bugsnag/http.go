package bugsnag

import (
	"net/http"

	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"
)

// HTTPRequestFormHook attaches an HTTP Request's Form to the request tab
func HTTPRequestFormHook(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
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
