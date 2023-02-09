package bugsnag

import (
	"errors"

	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"

	"github.com/Shopify/goose/v2/metrics"
)

func ErrorDetailsHook(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
	var errWithDetails interface {
		ErrorDetails() map[string]interface{}
	}
	if errors.As(event.Error.Err, &errWithDetails) {
		event.MetaData.Update(bugsnaggo.MetaData{
			"error": {
				"details": errWithDetails.ErrorDetails(),
			},
		})
	}
	return nil
}

func TemporaryErrorHook(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
	var tmpError interface {
		Temporary() bool
	}
	if errors.As(event.Error.Err, &tmpError) {
		temporary := tmpError.Temporary()
		event.Ctx = metrics.WithTag(event.Ctx, "error_temporary", temporary)
		event.MetaData.Update(bugsnaggo.MetaData{
			"error": {
				"temporary": temporary,
			},
		})
	}
	return nil
}

func TimeoutErrorHook(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
	var tmpError interface {
		Timeout() bool
	}
	if errors.As(event.Error.Err, &tmpError) {
		timeout := tmpError.Timeout()
		event.Ctx = metrics.WithTag(event.Ctx, "error_timeout", timeout)
		event.MetaData.Update(bugsnaggo.MetaData{
			"error": {
				"timeout": timeout,
			},
		})
	}
	return nil
}
