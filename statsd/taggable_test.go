package statsd

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var exampleBackend = NewForwardingBackend(func(_ context.Context, mType string, name string, value interface{}, tags []string, _ float64) error {
	_, err := fmt.Printf("%s: %s: %v %v\n", mType, name, value, tags)
	return err
})

type testTaggable Tags

func (t *testTaggable) StatsTags() Tags {
	return Tags(*t)
}

func ExampleWithTags() {
	SetBackend(exampleBackend)

	ctx := context.Background()
	ctx = WithTags(ctx, Tags{"user": "anonymous", "email": "unknown"})

	metric := &Counter{Name: "page.view"}
	metric.Count(ctx, 10)

	// Output:
	// count: page.view: 10 [email:unknown user:anonymous]
}

func ExampleWatchingTaggable() {
	SetBackend(exampleBackend)

	session := &testTaggable{"user": "anonymous", "email": "unknown"}

	ctx := context.Background()
	ctx = WatchingTaggable(ctx, session)

	metric := &Counter{Name: "page.view"}
	metric.Count(ctx, 10)

	// Output:
	// count: page.view: 10 [email:unknown user:anonymous]
}

func TestEmptyContext(t *testing.T) {
	ctx := context.Background()
	// Using a basic type on purpose, disable linter
	ctx = context.WithValue(ctx, "a", "b") //nolint:golint,staticcheck
	// Not showing up in tags
	assert.Empty(t, getStatsTags(ctx))
}

func TestWithTags(t *testing.T) {
	// Test that passing nil doesn't actually crash it, disable the linter
	ctx := WithTags(nil, Tags{"a": "b", "c": "d"}) //nolint:golint,staticcheck
	ctx = WithTags(ctx, Tags{"a": "e", "f": "g"})

	// Test it doesn't override
	assert.Equal(t, []string{
		"a:e",
		"c:d",
		"f:g",
	}, getStatsTags(ctx))
}

func TestWithTags_keyClash(t *testing.T) {
	ctx := context.Background()
	ctx = WithTags(ctx, Tags{"a": "b"})

	// tagsKey is an int declared as a contextKey, so trying to set an int shouldn't override the contextKey
	// Using a basic type on purpose, disable linter
	ctx = context.WithValue(ctx, int(tagsKey), "foo") //nolint:golint,staticcheck

	assert.Equal(t, []string{
		"a:b",
	}, getStatsTags(ctx))
}

func TestChildContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithTags(ctx, Tags{"a": "b"})
	ctx = WithTags(ctx, Tags{"c": "d"})
	ctx2 := WithTags(ctx, Tags{"c": "e"})

	// Test it doesn't override
	assert.Equal(t, []string{
		"a:b",
		"c:e",
	}, getStatsTags(ctx2))

	// Original still intact
	assert.Equal(t, []string{
		"a:b",
		"c:d",
	}, getStatsTags(ctx))
}

func TestWithTaggable(t *testing.T) {
	ctx := context.Background()
	ctx = WithTags(ctx, Tags{"a": "1"})

	l := &testTaggable{"a": 2, "b": 2}
	ctx = WithTaggable(ctx, l)
	ctx2 := WatchingTaggable(ctx, l)

	assert.Equal(t, []string{
		"a:2",
		"b:2",
	}, getStatsTags(ctx))

	(*l)["b"] = 3

	// Doesn't update
	assert.Equal(t, []string{
		"a:2",
		"b:2",
	}, getStatsTags(ctx))

	// But the context with WatchingTaggable does
	assert.Equal(t, []string{
		"a:2",
		"b:3",
	}, getStatsTags(ctx2))
}

func TestWatchingTaggable(t *testing.T) {
	ctx := context.Background()
	ctx = WithTags(ctx, Tags{"a": "1"})
	ctx = WithTags(ctx, Tags{"b": "1"})
	ctx2 := WatchingTaggable(ctx, &testTaggable{"a": 2, "c": 2})

	provider := &testTaggable{"a": 3}
	ctx3 := WatchingTaggable(ctx2, provider)

	assert.Equal(t, []string{
		"a:3",
		"b:1",
		"c:2",
	}, getStatsTags(ctx3))

	// Modification after call to WatchingLoggable
	(*provider)["a"] = 4
	assert.Equal(t, []string{
		"a:4",
		"b:1",
		"c:2",
	}, getStatsTags(ctx3))

	// New list also gets picked up
	*provider = testTaggable{"a": 5}
	assert.Equal(t, []string{
		"a:5",
		"b:1",
		"c:2",
	}, getStatsTags(ctx3))

	// Original contexts are untouched
	assert.Equal(t, []string{
		"a:2",
		"b:1",
		"c:2",
	}, getStatsTags(ctx2))

	assert.Equal(t, []string{
		"a:1",
		"b:1",
	}, getStatsTags(ctx))
}
