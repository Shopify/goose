package metrics

import (
	"context"

	"github.com/sirupsen/logrus"
)

var _ fieldLogger = (*logrus.Logger)(nil)
var _ fieldLogger = (*logrus.Entry)(nil)

type fieldLogger interface {
	logrus.FieldLogger
	WithContext(ctx context.Context) *logrus.Entry
}

func NewStandardLogrusBackend(level logrus.Level) Backend {
	return NewLogrusBackend(logrus.StandardLogger(), level)
}

// NewLogrusBackend creates a new Backend that emits statsd metrics to the logger.
//
// namespace is a global namespace to apply to every metric emitted.
// tags is a global set of tags that will be added to every metric emitted.
func NewLogrusBackend(logger fieldLogger, level logrus.Level) Backend {
	c := &logrusBackend{
		logger: logger,
		level:  level,
	}
	return NewForwardingBackend(c.log)
}

type logrusBackend struct {
	logger fieldLogger
	level  logrus.Level
}

func (c *logrusBackend) log(ctx context.Context, metric *Metric) error {
	c.logger.
		WithContext(ctx).
		WithField("metric", metric).
		Log(c.level, "emit metric")
	return nil
}
