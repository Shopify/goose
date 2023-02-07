package metrics

import (
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func NewExampleBackend() Backend {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.JSONFormatter{
		DisableTimestamp: true,
	})

	backend := NewLogrusBackend(logger, logrus.InfoLevel)
	backend = BackendWithDefaultWrappers(backend, "example")
	return backend
}

type testTaggable Tags

func (t *testTaggable) StatsTags() Tags {
	return Tags(*t)
}

type testLoggable logrus.Fields

func (l *testLoggable) LogFields() logrus.Fields {
	return logrus.Fields(*l)
}

func (l *testLoggable) StatsTags() Tags {
	return SelectKeys(l.LogFields(), "testField")
}

func ExampleWithTags() {
	SetDefaultBackend(NewExampleBackend())

	ctx := context.Background()
	ctx = WithTags(ctx, Tags{"user": "anonymous", "email": "unknown"})

	metric := &Counter{Name: "page.view"}
	metric.Count(ctx, 10)

	// Output:
	// {"level":"info","metric":{"Type":"Count","Name":"example.page.view","Value":10,"Tags":{"email":"unknown","user":"anonymous"},"Rate":1},"msg":"emit metric"}
}

func ExampleWatchingTaggable() {
	SetDefaultBackend(NewExampleBackend())

	session := &testTaggable{"user": "anonymous", "email": "unknown"}

	ctx := context.Background()
	ctx = WatchingTaggable(ctx, session)

	metric := &Counter{Name: "page.view"}
	metric.Count(ctx, 10)

	// Output:
	// {"level":"info","metric":{"Type":"Count","Name":"example.page.view","Value":10,"Tags":{"email":"unknown","user":"anonymous"},"Rate":1},"msg":"emit metric"}
}

func TestEmptyContext(t *testing.T) {
	ctx := context.Background()
	// Using a basic type on purpose, disable linter
	ctx = context.WithValue(ctx, "a", "b") //nolint:revive,staticcheck
	// Not showing up in tags
	assert.Empty(t, TagsFromContext(ctx))
}

func TestWithTags(t *testing.T) {
	ctx := context.Background()
	ctx = WithTags(ctx, Tags{"a": "b", "c": "d"})
	ctx = WithTags(ctx, Tags{"a": "e", "f": "g"})

	// Test it doesn't override
	assert.Equal(t, Tags{
		"a": "e",
		"c": "d",
		"f": "g",
	}, TagsFromContext(ctx))
}

func TestChildContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithTags(ctx, Tags{"a": "b"})
	ctx = WithTags(ctx, Tags{"c": "d"})
	ctx2 := WithTags(ctx, Tags{"c": "e"})

	// Test it doesn't override
	assert.Equal(t, Tags{
		"a": "b",
		"c": "e",
	}, TagsFromContext(ctx2))

	// Original still intact
	assert.Equal(t, Tags{
		"a": "b",
		"c": "d",
	}, TagsFromContext(ctx))
}

func TestWithTaggable(t *testing.T) {
	ctx := context.Background()
	ctx = WithTags(ctx, Tags{"a": 1})

	l := &testTaggable{"a": 2, "b": 2}
	ctx = WithTaggable(ctx, l)
	ctx2 := WatchingTaggable(ctx, l)

	assert.Equal(t, Tags{
		"a": 2,
		"b": 2,
	}, TagsFromContext(ctx))

	(*l)["b"] = 3

	// Doesn't update
	assert.Equal(t, Tags{
		"a": 2,
		"b": 2,
	}, TagsFromContext(ctx))

	// But the context with WatchingTaggable does
	assert.Equal(t, Tags{
		"a": 2,
		"b": 3,
	}, TagsFromContext(ctx2))
}

func TestWatchingTaggable(t *testing.T) {
	ctx := context.Background()
	ctx = WithTags(ctx, Tags{"a": 1})
	ctx = WithTags(ctx, Tags{"b": 1})
	ctx2 := WatchingTaggable(ctx, &testTaggable{"a": 2, "c": 2})

	provider := &testTaggable{"a": 3}
	ctx3 := WatchingTaggable(ctx2, provider)

	assert.Equal(t, Tags{
		"a": 3,
		"b": 1,
		"c": 2,
	}, TagsFromContext(ctx3))

	// Modification after call to WatchingLoggable
	(*provider)["a"] = 4
	assert.Equal(t, Tags{
		"a": 4,
		"b": 1,
		"c": 2,
	}, TagsFromContext(ctx3))

	// New list also gets picked up
	*provider = testTaggable{"a": 5}
	assert.Equal(t, Tags{
		"a": 5,
		"b": 1,
		"c": 2,
	}, TagsFromContext(ctx3))

	// Original contexts are untouched
	assert.Equal(t, Tags{
		"a": 2,
		"b": 1,
		"c": 2,
	}, TagsFromContext(ctx2))

	assert.Equal(t, Tags{
		"a": 1,
		"b": 1,
	}, TagsFromContext(ctx))
}
