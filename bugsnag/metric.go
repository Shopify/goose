package bugsnag

import (
	bugsnaggo "github.com/bugsnag/bugsnag-go/v2"

	"github.com/Shopify/goose/v2/metrics"
)

var reportMetric = &metrics.Counter{Name: "bugsnag.report"}

func NewMetricsCounterHook(counter *metrics.Counter) func(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
	return func(event *bugsnaggo.Event, config *bugsnaggo.Configuration) error {
		counter.Incr(event.Ctx)
		return nil
	}
}
