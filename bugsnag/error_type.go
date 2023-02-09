package bugsnag

import (
	"errors"
	"reflect"

	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"

	"github.com/Shopify/goose/v2/metrics"
)

func rootError(err error) error {
	for {
		var unwrapped error

		var wrappedError interface{ Unwrap() error }
		if errors.As(err, &wrappedError) {
			unwrapped = wrappedError.Unwrap()
		}

		var causedError interface{ Cause() error }
		if unwrapped == nil && errors.As(err, &causedError) {
			unwrapped = causedError.Cause()
		}

		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}

func ErrorTypeHook(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
	if event.Error != nil {
		err := rootError(event.Error.Err)
		errorType := reflect.TypeOf(err)

		if errorType.Kind() == reflect.Pointer {
			errorType = errorType.Elem()
		}
		var typeStr string
		pkgPath := errorType.PkgPath()
		if pkgPath == "errors" || pkgPath == "github.com/pkg/errors" {
			typeStr = "basic_error" // Don't expose the internals of the errors packages
		} else {
			typeStr = errorType.String()
		}

		event.Ctx = metrics.WithTag(event.Ctx, "error_type", typeStr)
		event.MetaData.Update(bugsnaggo.MetaData{
			"error": {
				"type": typeStr,
			},
		})
	}
	return nil
}
