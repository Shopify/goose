package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

type logFieldsKeyType struct{}

var (
	logFieldsKey = logFieldsKeyType{}
)

type Loggable interface {
	LogFields() logrus.Fields
}

type keyValueContext struct {
	context.Context
	key   string
	value interface{}
}

func (c *keyValueContext) Value(key interface{}) interface{} {
	if key == logFieldsKey {
		prev := GetLoggableValues(c.Context)
		prev[c.key] = c.value
		return prev
	}
	return c.Context.Value(key)
}

type loggableContext struct {
	context.Context
	loggable Loggable
}

func (c *loggableContext) Value(key interface{}) interface{} {
	if key == logFieldsKey {
		prev := GetLoggableValues(c.Context)
		for k, v := range c.loggable.LogFields() {
			prev[k] = v
		}
		return prev
	}
	return c.Context.Value(key)
}

// WithField attaches a key-value pair to a Context.
// Upon logging, the pair will be attached as metadata.
func WithField(ctx context.Context, k string, v interface{}) context.Context {
	return &keyValueContext{Context: ctx, key: k, value: v}
}

// WithFields attaches fields to a Context.
// Upon logging, those fields will be attached as metadata.
func WithFields(ctx context.Context, m logrus.Fields) context.Context {
	for k, v := range m {
		ctx = WithField(ctx, k, v)
	}
	return ctx
}

// WithLoggable attaches a Loggable's fields to the Context.
// Upon logging, those fields will be attached as metadata.
func WithLoggable(ctx context.Context, l Loggable) context.Context {
	return WithFields(ctx, l.LogFields())
}

// WatchingLoggable attaches a Loggable to the Context.
// Upon logging, LogFields will be called and the fields will be attached as metadata.
func WatchingLoggable(ctx context.Context, l Loggable) context.Context {
	return &loggableContext{Context: ctx, loggable: l}
}

// GetLoggableValue returns the value of the metadata currently attached to the Context.
func GetLoggableValue(ctx context.Context, key string) interface{} {
	fields := GetLoggableValues(ctx)
	if v, ok := fields[key]; ok {
		return v
	}
	return nil
}

func GetLoggableValues(ctx context.Context) logrus.Fields {
	if ctx != nil {
		fields, _ := ctx.Value(logFieldsKey).(logrus.Fields)
		if fields != nil {
			return fields
		}
	}

	return logrus.Fields{}
}
