package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

// This create a private key-space in the Context, meaning that only this package can get or set "contextKey" types
type contextKey int

const (
	logFieldsKey contextKey = iota
)

type Loggable interface {
	LogFields() logrus.Fields
}

// Valuer is essentially a Context, but down-scoped to what Loggable expects
type Valuer interface {
	Value(key interface{}) interface{}
}

type keyValueContext struct {
	context.Context
	key   string
	value interface{}
}

func (c *keyValueContext) Value(key interface{}) interface{} {
	if key == logFieldsKey {
		prev := getLoggableValues(c.Context)
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
		prev := getLoggableValues(c.Context)
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
func GetLoggableValue(ctx Valuer, key string) interface{} {
	fields := getLoggableValues(ctx)
	if v, ok := fields[key]; ok {
		return v
	}
	return nil
}

func getLoggableValues(ctx Valuer) logrus.Fields {
	if ctx != nil {
		fields, _ := ctx.Value(logFieldsKey).(logrus.Fields)
		if fields != nil {
			return fields
		}
	}

	return logrus.Fields{}
}
