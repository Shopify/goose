package logger

import (
	"context"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func ExampleWithField() {
	ctx := context.Background()
	ctx = WithField(ctx, "foo", "bar")

	log := New("testing")

	// Typically called as log(ctx, nil).Debug("the message")
	entry := log(ctx, nil)
	fmt.Printf("%+v", entry.Data["foo"])

	// Output:
	// bar
}

func ExampleWithFields() {
	ctx := context.Background()
	ctx = WithFields(ctx, logrus.Fields{
		"foo": "bar",
	})

	log := New("testing")

	// Typically called as log(ctx, nil).Debug("the message")
	entry := log(ctx, nil)
	fmt.Printf("%+v", entry.Data["foo"])

	// Output:
	// bar
}

type exampleLoggable logrus.Fields

func (l exampleLoggable) LogFields() logrus.Fields {
	return logrus.Fields(l)
}

func ExampleWatchingLoggable() {
	loggable := &exampleLoggable{"foo": "bar"}

	ctx := context.Background()
	ctx = WatchingLoggable(ctx, loggable)

	log := New("testing")

	// Typically called as log(ctx, nil).Debug("the message")
	entry := log(ctx, nil)
	fmt.Printf("%+v", entry.Data["foo"])

	// Output:
	// bar
}

func TestEmptyContext(t *testing.T) {
	ctx := context.Background()
	// Using a basic type on purpose, disable linter
	ctx = context.WithValue(ctx, "a", "b") // nolint: golint
	// Not showing up in logs
	checkData(ctx, t, logrus.Fields{"component": "testing"})
}

func TestWithFields(t *testing.T) {
	// Test that passing nil doesn't actually crash it, disable the linter
	ctx := WithFields(nil, logrus.Fields{"a": "b", "c": "d"}) //nolint: staticcheck
	ctx = WithFields(ctx, logrus.Fields{"a": "e", "f": "g"})

	// Test overrides
	checkData(ctx, t, logrus.Fields{
		"a": "e",
		"c": "d",
		"f": "g",
	})
}

func TestWithLoggable(t *testing.T) {
	ctx := context.Background()
	l := &testLoggable{
		keys: logrus.Fields{"a": "1", "b": "2"},
	}
	ctx = WithLoggable(ctx, l)

	checkData(ctx, t, logrus.Fields{
		"a": "1",
		"b": "2",
	})

	l.keys["b"] = "3"

	// Hasn't changed
	checkData(ctx, t, logrus.Fields{
		"a": "1",
		"b": "2",
	})
}

func TestWithField(t *testing.T) {
	// Test that passing nil doesn't actually crash it, disable the linter
	ctx := WithField(nil, "a", "b") //nolint: staticcheck
	ctx = WithField(ctx, "c", "d")
	ctx = WithField(ctx, "c", "e")

	// Test overrides
	checkData(ctx, t, logrus.Fields{
		"a": "b",
		"c": "e",
	})
}

func TestLoggableKeyClash(t *testing.T) {
	ctx := context.Background()
	ctx = WithField(ctx, "a", "b")

	// logFieldsKey is an int declared as a contextKey, so trying to set an int shouldn't override the contextKey
	// Using a basic type on purpose, disable linter
	ctx = context.WithValue(ctx, int(logFieldsKey), "foo") //nolint: golint

	checkData(ctx, t, logrus.Fields{
		"a": "b",
	})
}

func TestChildContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithField(ctx, "a", "b")
	ctx = WithField(ctx, "c", "d")
	ctx2 := WithField(ctx, "c", "e")

	// Test overrides
	checkData(ctx2, t, logrus.Fields{
		"a": "b",
		"c": "e",
	})

	// Original still intact
	checkData(ctx, t, logrus.Fields{
		"a": "b",
		"c": "d",
	})
}

type testLoggable struct {
	keys logrus.Fields
}

func (l *testLoggable) LogFields() logrus.Fields {
	return l.keys
}

func TestWatchingLoggable(t *testing.T) {
	ctx := context.Background()
	ctx = WithField(ctx, "a", "1")
	ctx = WithField(ctx, "b", "1")

	ctx2 := WatchingLoggable(ctx, &testLoggable{
		keys: logrus.Fields{"a": "2", "c": "2"},
	})

	loggable := &testLoggable{
		keys: logrus.Fields{"a": "3"},
	}
	ctx3 := WatchingLoggable(ctx2, loggable)

	checkData(ctx3, t, logrus.Fields{
		"a": "3",
		"b": "1",
		"c": "2",
	})

	// Modification after call to WatchingLoggable
	loggable.keys["a"] = "4"
	checkData(ctx3, t, logrus.Fields{
		"a": "4",
		"b": "1",
		"c": "2",
	})

	// New map also gets picked up
	loggable.keys = logrus.Fields{"a": "5"}
	checkData(ctx3, t, logrus.Fields{
		"a": "5",
		"b": "1",
		"c": "2",
	})

	// Original contexts are untouched
	checkData(ctx2, t, logrus.Fields{
		"a": "2",
		"b": "1",
		"c": "2",
	})

	checkData(ctx, t, logrus.Fields{
		"a": "1",
		"b": "1",
	})
}

func checkData(ctx context.Context, t *testing.T, expected logrus.Fields) {
	expected["component"] = "testing" // Will always be there
	assert.Equal(t, expected, New("testing")(ctx, nil).Data)
}

func TestGetLoggableValue(t *testing.T) {
	ctx := context.Background()
	ctx = WithField(ctx, "foo", "bar")

	value := GetLoggableValue(ctx, "foo")
	assert.Equal(t, "bar", value)
}

func TestGetLoggableValues(t *testing.T) {
	ctx := context.Background()
	ctx = WithField(ctx, "foo", "bar")
	ctx = WithField(ctx, "poipoi", true)

	values := GetLoggableValues(ctx)
	assert.Equal(t, logrus.Fields{"foo": "bar", "poipoi": true}, values)
}
