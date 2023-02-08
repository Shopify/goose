package metrics

import (
	"context"

	"github.com/Shopify/goose/v2/logger"
)

// Taggable is meant to be attached to a Context, such that StatsTags() will be appended to recorded metrics.
type Taggable interface {
	StatsTags() Tags
}

// WithTaggable attaches a Taggable's tags to a Context.
// When a metric is recorded, those tags will be appended.
func WithTaggable(ctx context.Context, t Taggable) context.Context {
	return WithTags(ctx, t.StatsTags())
}

// taggableContext is the same as tagContext, but with dynamic tags.
type taggableContext struct {
	context.Context
	taggable Taggable
}

func (c *taggableContext) Value(key interface{}) interface{} {
	if key == tagsContextKey {
		tags := TagsFromContext(c.Context)
		return tags.Merge(c.taggable.StatsTags())
	}
	return c.Context.Value(key)
}

// WatchingTaggable attaches a Taggable to a Context.
// When a metric is recorded, StatsTags() will be called and the tags will be appended.
func WatchingTaggable(ctx context.Context, t Taggable) context.Context {
	return &taggableContext{Context: ctx, taggable: t}
}

// WithTagLoggable combines WithTaggable and logger.WithLoggable
// If the Loggable is a Taggable already (implements StatsTags), it will be used directly.
// If StatsTags doesn't exist, LogFields() will be used instead.
func WithTagLoggable(ctx context.Context, l logger.Loggable) context.Context {
	ctx = logger.WithLoggable(ctx, l)
	if taggable, ok := l.(Taggable); ok {
		ctx = WithTaggable(ctx, taggable)
	} else {
		ctx = WithTaggable(ctx, tagLoggable{l})
	}
	return ctx
}

type tagLoggable struct {
	logger.Loggable
}

func (l tagLoggable) StatsTags() Tags {
	return Tags(l.LogFields())
}

// WatchingTagLoggable combines WatchingTaggable and logger.WatchingLoggable
// If the Loggable is a Taggable already (implements StatsTags), it will be used directly.
// If StatsTags doesn't exist, LogFields() will be used instead.
func WatchingTagLoggable(ctx context.Context, l logger.Loggable) context.Context {
	ctx = logger.WatchingLoggable(ctx, l)
	if taggable, ok := l.(Taggable); ok {
		ctx = WatchingTaggable(ctx, taggable)
	} else {
		ctx = WatchingTaggable(ctx, tagLoggable{l})
	}
	return ctx
}

// SelectKeys returns a map containing only the specified fields
// This argument purposefully not typed as Tags, such that logrus.Fields and Tags can both be passed without additional casting.
// Useful when specifying a Loggable/Taggable
func SelectKeys(m map[string]interface{}, keys ...string) Tags {
	tags := Tags{}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			tags[k] = v
		}
	}
	return tags
}
