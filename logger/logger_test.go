package logger

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func withStdoutLogger(ctx context.Context) context.Context {
	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Formatter = &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	}

	RegisterHook(logger)
	return WithLogger(ctx, logger)
}

func withBufLogger(ctx context.Context) (context.Context, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	logger := logrus.New()
	logger.Out = buf
	logger.Formatter = &logrus.TextFormatter{
		DisableColors:    true,
		DisableTimestamp: true,
	}

	RegisterHook(logger)
	return WithLogger(ctx, logger), buf
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
	ctx := context.Background()
	log := New("foo")

	t.Run("with err", func(t *testing.T) {
		ctx, buf := withBufLogger(ctx)
		err := errors.New("bad stuff")
		log(ctx, err).WithField("a", "b").Info("test")

		assert.Equal(t, "level=info msg=test a=b component=foo error=\"bad stuff\"\n", buf.String())
	})

	t.Run("err with fields", func(t *testing.T) {
		ctx, buf := withBufLogger(ctx)
		cause := &logFieldsErr{"bad stuff", logrus.Fields{"foo": "bar", "baz": "qux"}}
		err := errors.Wrap(cause, "cause")
		log(ctx, err).WithField("a", "b").Info("test")

		assert.Equal(t, "level=info msg=test a=b baz=qux component=foo error=\"cause: bad stuff\" foo=bar\n", buf.String())
	})

	t.Run("without err", func(t *testing.T) {
		ctx, buf := withBufLogger(ctx)
		log(ctx).WithField("a", "b").Info("test")
		assert.Equal(t, "level=info msg=test a=b component=foo\n", buf.String())
	})

	t.Run("with 2 errors", func(t *testing.T) {
		ctx, buf := withBufLogger(ctx)
		err1 := errors.New("bad stuff")
		err2 := errors.New("also bad stuff")
		log(ctx, err1, err2).WithField("a", "b").Info("test")

		assert.Equal(t, "level=info msg=test a=b component=foo error=\"bad stuff\\nalso bad stuff\"\n", buf.String())
	})
}

func TestContextLog(t *testing.T) {
	origGlobal := logrus.Fields{}
	for k, v := range GlobalFields {
		origGlobal[k] = v
	}
	defer func() { GlobalFields = origGlobal }()

	log := New("foo")

	GlobalFields["testKey"] = "value"

	ctx := context.Background()
	ctx, buf := withBufLogger(ctx)

	ctx = WithField(ctx, "bar", "baz")
	log(ctx).WithField("a", "b").Info("test")
	assert.Equal(t, "level=info msg=test a=b bar=baz component=foo testKey=value\n", buf.String())
}

func TestLogIfError(t *testing.T) {
	ctx := context.Background()

	t.Run("no error", func(t *testing.T) {
		ctx, buf := withBufLogger(ctx)
		fn := func() error { return nil }
		LogIfError(ctx, fn, nil, "")
		assert.Equal(t, "", buf.String())
	})

	t.Run("error", func(t *testing.T) {
		ctx, buf := withBufLogger(ctx)
		fn := func() error { return errors.New("foo") }
		LogIfError(ctx, fn, nil, "msg")
		assert.Equal(t, "level=error msg=msg error=foo\n", buf.String())
	})

	t.Run("error with context field", func(t *testing.T) {
		ctx, buf := withBufLogger(ctx)
		fn := func() error { return errors.New("foo") }
		ctx = WithField(ctx, "test", "bar")
		LogIfError(ctx, fn, nil, "msg")
		assert.Equal(t, "level=error msg=msg error=foo test=bar\n", buf.String())
	})
}
