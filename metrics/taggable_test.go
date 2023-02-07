package metrics

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/Shopify/goose/v2/logger"
)

var exampleBackend = NewForwardingBackend(func(_ context.Context, mType string, name string, value interface{}, tags Tags, _ float64) error {
	_, err := fmt.Printf("%s: %s: %v %v\n", mType, name, value, tags)
	return err
})

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
	SetBackend(exampleBackend)

	ctx := context.Background()
	ctx = WithTags(ctx, Tags{"user": "anonymous", "email": "unknown"})

	metric := &Counter{Name: "page.view"}
	metric.Count(ctx, 10)

	// Output:
	// count: page.view: 10 map[email:unknown user:anonymous]
}

func ExampleWatchingTaggable() {
	SetBackend(exampleBackend)

	session := &testTaggable{"user": "anonymous", "email": "unknown"}

	ctx := context.Background()
	ctx = WatchingTaggable(ctx, session)

	metric := &Counter{Name: "page.view"}
	metric.Count(ctx, 10)

	// Output:
	// count: page.view: 10 map[email:unknown user:anonymous]
}

func ExampleSelectKeys() {
	// <setup for example>
	logrusLogger := logrus.New()
	logrusLogger.Out = os.Stdout
	logrusLogger.Formatter = &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	}
	entry := logrus.NewEntry(logrusLogger)
	SetBackend(exampleBackend)
	// </setup for example>

	session := &testLoggable{"foo": "bar", "testField": "test"}

	ctx := context.Background()
	ctx = WithTagLoggable(ctx, session)

	metric := &Counter{Name: "page.view"}
	metric.Count(ctx, 10)

	logger.ContextLog(ctx, nil, entry).Info("example")

	// Output:
	// count: page.view: 10 map[testField:test]
	// level=info msg=example foo=bar testField=test
}

func TestEmptyContext(t *testing.T) {
	ctx := context.Background()
	// Using a basic type on purpose, disable linter
	ctx = context.WithValue(ctx, "a", "b") //nolint:revive,staticcheck
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
