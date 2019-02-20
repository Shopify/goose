package logger

import (
	"bytes"
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func buildLogger() (Logger, *bytes.Buffer) {
	buf := bytes.NewBuffer(nil)
	logrusLogger := logrus.New()
	logrusLogger.Out = buf
	logrusLogger.Formatter = &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	}
	entry := logrus.NewEntry(logrusLogger)

	logger := func(ctx Valuer, err ...error) *logrus.Entry {
		return ContextLog(ctx, err, entry)
	}

	return logger, buf
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

		assert.Equal(t, logrus.Fields{
			"component": "foo",
			"a":         "b",
			"error":     err1,
			"error1":    err2,
		}, entry.Data)
	})

	t.Run("with several errors", func(t *testing.T) {
		logger := New("foo")
		ctx := context.Background()
		err1 := errors.Wrap(errors.New("err1"), "wrapped1")
		err2 := errors.New("err2")
		err3 := errors.New("err3")
		err4 := errors.Wrap(errors.New("err4"), "wrapped4")
		err5 := errors.New("err5")

		entry := logger(ctx, err1, err2, err3, err4, err5).WithField("a", "b")

		assert.Equal(t, logrus.Fields{
			"component": "foo",
			"a":         "b",
			"error":     err1,
			"cause":     errors.Cause(err1),
			"error1":    err2,
			"error2":    err3,
			"error3":    err4,
			"cause3":    errors.Cause(err4),
			"error4":    err5,
		}, entry.Data)
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

	ctx := context.Background()
	ctx = WithField(ctx, "bar", "baz")
	entry := logger(ctx, nil).WithField("a", "b")
	assert.Equal(t, logrus.Fields{
		"component": "foo",
		"bar":       "baz",
		"a":         "b",
		"testKey":   "value",
	}, entry.Data)
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
	{
		logger, buf := buildLogger()
		nestedFn := func() error { return errors.New("root_cause") }
		fn := func() error {
			err := nestedFn()
			return errors.Wrap(err, "err_msg")
		}
		LogIfError(ctx, fn, logger, "msg")
		assert.Equal(t, "level=error msg=msg cause=root_cause error=\"err_msg: root_cause\"\n", buf.String())
	}
}
