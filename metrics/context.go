package metrics

import (
	"context"

	"github.com/Shopify/goose/v2/logger"
)

// contextKey is a private key-space in the Context, meaning that only this package can get or set "contextKey" types
// This create a private key-space in the Context, meaning that only this package can get or set "contextKey" types
type tagsContextKeyType struct{}

type backendContextKeyType struct{}

var (
	backendContextKey = backendContextKeyType{}
	tagsContextKey    = tagsContextKeyType{}
)

func ContextWithBackend(ctx context.Context, backend Backend) context.Context {
	return context.WithValue(ctx, backendContextKey, backend)
}

func BackendFromContext(ctx context.Context) Backend {
	backend, _ := ctx.Value(backendContextKey).(Backend)
	if backend != nil {
		return backend
	}
	return DefaultBackend()
}

func TagsFromContext(ctx context.Context) Tags {
	if tags, ok := ctx.Value(tagsContextKey).(Tags); ok {
		return tags
	}
	return Tags{}
}

// tagContext wraps a parent Context and override the behaviour of Value(), similar to how context.valueCtx works.
// The difference with context.valueCtx is that it _appends_ to the parent's value instead of replacing it.
type tagContext struct {
	context.Context
	key   string
	value interface{}
}

func (c *tagContext) Value(key interface{}) interface{} {
	if key == tagsContextKey {
		prev := TagsFromContext(c.Context)
		prev[c.key] = c.value
		return prev
	}
	return c.Context.Value(key)
}

// WithTag attaches a key-value pair to a Context.
// Upon recording a metric, the pair will be attached as a tag.
func WithTag(ctx context.Context, key string, value interface{}) context.Context {
	return &tagContext{Context: ctx, key: key, value: value}
}

// WithTags attaches a map of tags to a Context
// Upon recording a metric, those fields will be attached as tags.
func WithTags(ctx context.Context, tags Tags) context.Context {
	for k, v := range tags {
		ctx = WithTag(ctx, k, v)
	}
	return ctx
}

// WithTagLogFields combines logger.WithFields(fields) and WithTagMap(tags)
// This simplifies the common operation of adding fields to the logger and the metrics
// This argument purposefully not typed as Tags, such that logrus.Fields and Tags can both be passed without additional casting.
func WithTagLogFields(ctx context.Context, tags map[string]interface{}) context.Context {
	ctx = logger.WithFields(ctx, tags)
	ctx = WithTags(ctx, tags)
	return ctx
}
