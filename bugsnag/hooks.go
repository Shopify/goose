package bugsnag

import (
	"context"

	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"

	"github.com/Shopify/goose/v2/logger"
)

func RegisterDefaultHooks() {
	for _, hook := range DefaultHooks {
		bugsnaggo.OnBeforeNotify(hook)
	}
}

type Hook func(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error

var DefaultHooks = []Hook{
	GlobalHook.Hook,
	NewMetricsCounterHook(reportMetric),
	ErrorClassHook,
	ErrorTypeHook,
	ErrorDetailsHook,
	TemporaryErrorHook,
	TimeoutErrorHook,
	ContextFieldsHook,
	TabHook,
	LogFieldsHook,
}

func ContextFieldsHook(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
	if event.Ctx == nil {
		event.Ctx = context.Background() // Avoid nil pointers
	} else {
		event.MetaData.Update(bugsnaggo.MetaData{
			"metadata": logger.GetLoggableValues(event.Ctx),
		})
	}
	return nil
}
