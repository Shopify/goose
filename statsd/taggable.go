package statsd

import (
	"context"
	"fmt"
	"sort"
)

// This create a private key-space in the Context, meaning that only this package can get or set "contextKey" types
type contextKey int

const (
	tagsKey contextKey = iota
)

type Tags map[string]interface{}

// Taggable is meant to be attached to a Context, such that StatsTags() will be appended to recorded metrics.
type Taggable interface {
	StatsTags() Tags
}

// keyValueContext wraps a parent Context and override the behaviour of Value(), similar to how context.valueCtx works.
// The difference with context.valueCtx is that it _appends_ to the parent's value instead of replacing it.
type keyValueContext struct {
	context.Context
	key   string
	value interface{}
}

func (c *keyValueContext) Value(key interface{}) interface{} {
	if key == tagsKey {
		prev := getStatsTagsMap(c.Context)
		prev[c.key] = c.value
		return prev
	}
	return c.Context.Value(key)
}

// taggableContext is the same as tagContext, but with dynamic tags.
type taggableContext struct {
	context.Context
	taggable Taggable
}

func (c *taggableContext) Value(key interface{}) interface{} {
	if key == tagsKey {
		prev := getStatsTagsMap(c.Context)
		for k, v := range c.taggable.StatsTags() {
			prev[k] = v
		}
		return prev
	}
	return c.Context.Value(key)
}

// WithTag attaches a key-value pair to a Context.
// Upon recording a metric, the pair will be attached as a tag.
func WithTag(ctx context.Context, k string, v interface{}) context.Context {
	return &keyValueContext{Context: ctx, key: k, value: v}
}

// WithTags attaches fields to a Context.
// Upon recording a metric, those fields will be attached as tags.
func WithTags(ctx context.Context, t Tags) context.Context {
	for k, v := range t {
		ctx = WithTag(ctx, k, v)
	}
	return ctx
}

// WithTaggable attaches a Taggable's tags to a Context.
// When a metric is recorded, those tags will be appended.
func WithTaggable(ctx context.Context, t Taggable) context.Context {
	return WithTags(ctx, t.StatsTags())
}

// WatchingTaggable attaches a Taggable to a Context.
// When a metric is recorded, StatsTags() will be called and the tags will be appended.
func WatchingTaggable(ctx context.Context, t Taggable) context.Context {
	return &taggableContext{Context: ctx, taggable: t}
}

// getStatsTags returns the merged tags as a list
// Meant to be used by the metrics when inlining the tags
func getStatsTags(ctx context.Context, extraTagList ...Tags) []string {
	tags := getStatsTagsMap(ctx)
	for _, extraTags := range extraTagList {
		for k, v := range extraTags {
			tags[k] = v
		}
	}

	list := make([]string, 0, len(tags))
	for k, v := range tags {
		list = append(list, fmt.Sprintf("%s:%v", k, v))
	}

	sort.Strings(list)
	return list
}

// getStatsTagsMap returns the merged tags as a map
func getStatsTagsMap(ctx context.Context) Tags {
	if ctx != nil {
		fields, _ := ctx.Value(tagsKey).(Tags)
		if fields != nil {
			return fields
		}
	}

	return Tags{}
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
