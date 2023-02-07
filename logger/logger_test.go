package logger

import (
	"bytes"
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type contextKeyType string

func buildLogger() (Logger, *bytes.Buffer) {
	buf := bytes.NewBuffer(nil)
	logrusLogger := logrus.New()
	logrusLogger.Out = buf
	logrusLogger.Formatter = &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	}
	entry := logrus.NewEntry(logrusLogger)

	logger := func(ctx context.Context, err ...error) *logrus.Entry {
		return ContextLog(ctx, err, entry)
	}

	return logger, buf
}

type logFieldsErr struct {
	msg    string
	fields logrus.Fields
}

func (e *logFieldsErr) LogFields() logrus.Fields {
	return e.fields
}

func (e *logFieldsErr) Error() string {
	return e.msg
}

func TestNew_OptionalErr(t *testing.T) {
	origGlobal := logrus.Fields{}
	for k, v := range GlobalFields {
		origGlobal[k] = v
	}
	defer func() { GlobalFields = origGlobal }()

	t.Run("with err", func(t *testing.T) {
		logger := New("foo")
		ctx := context.Background()
		err := errors.New("bad stuff")

		entry := logger(ctx, err).WithField("a", "b")

		assert.Equal(t, logrus.Fields{
			"component": "foo",
			"a":         "b",
			"error":     err,
		}, entry.Data)
	})

	t.Run("err with fields", func(t *testing.T) {
		logger := New("foo")
		ctx := context.Background()
		cause := &logFieldsErr{"bad stuff", logrus.Fields{"foo": "bar", "baz": "qux"}}
		err := errors.Wrap(cause, "cause")

		entry := logger(ctx, err).WithField("a", "b")

		assert.Equal(t, logrus.Fields{
			"component": "foo",
			"a":         "b",
			"foo":       "bar",
			"baz":       "qux",
			"error":     err,
		}, entry.Data)
	})

	t.Run("without err", func(t *testing.T) {
		logger := New("foo")
		ctx := context.Background()
		entry := logger(ctx).WithField("a", "b")
		assert.Equal(t, logrus.Fields{
			"component": "foo",
			"a":         "b",
		}, entry.Data)
	})

	t.Run("with 2 errors", func(t *testing.T) {
		logger := New("foo")
		ctx := context.Background()
		err1 := errors.New("bad stuff")
		err2 := errors.New("also bad stuff")

		entry := logger(ctx, err1, err2).WithField("a", "b")

		assert.Equal(t, "b", entry.Data["a"])
		assert.EqualError(t, entry.Data["error"].(error), "bad stuff\nalso bad stuff")
	})
}

func TestContextLog(t *testing.T) {
	origGlobal := logrus.Fields{}
	for k, v := range GlobalFields {
		origGlobal[k] = v
	}
	defer func() { GlobalFields = origGlobal }()

	logger := New("foo")

	GlobalFields["testKey"] = "value"

	ctx := context.WithValue(context.Background(), contextKeyType("ctxValue"), "value")

	ctx = WithField(ctx, "bar", "baz")
	entry := logger(ctx, nil).WithField("a", "b")
	assert.Equal(t, logrus.Fields{
		"component": "foo",
		"bar":       "baz",
		"a":         "b",
		"testKey":   "value",
	}, entry.Data)
	assert.Equal(t, ctx, entry.Context)
}

func TestLogIfError(t *testing.T) {
	ctx := context.Background()
	{
		logger, buf := buildLogger()
		fn := func() error { return nil }
		LogIfError(ctx, fn, logger, "")
		assert.Equal(t, "", buf.String())
	}
	{
		logger, buf := buildLogger()
		fn := func() error { return errors.New("foo") }
		LogIfError(ctx, fn, logger, "msg")
		assert.Equal(t, "level=error msg=msg error=foo\n", buf.String())
	}
	{
		logger, buf := buildLogger()
		fn := func() error { return errors.New("foo") }
		ctx := WithField(ctx, "test", "bar")
		LogIfError(ctx, fn, logger, "msg")
		assert.Equal(t, "level=error msg=msg error=foo test=bar\n", buf.String())
	}
}
