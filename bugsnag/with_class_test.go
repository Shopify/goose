package bugsnag

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestExtractWithoutErrorClass(t *testing.T) {
	err := errors.New("test error")
	require.Equal(t, "test error", extractErrorClass(err))
}

func TestWithErrorClass(t *testing.T) {
	err := WithErrorClass(errors.New("test error"), "FOO")
	require.Equal(t, "FOO", extractErrorClass(err))
}

func TestWrappedWithErrorClass(t *testing.T) {
	err := errors.Wrapf(WithErrorClass(errors.New("test error"), "FOO"), "other %v", "bar")
	require.Equal(t, "FOO", extractErrorClass(err))
}

func TestFormatErrorClass(t *testing.T) {
	err := WithErrorClass(errors.New("test error"), "FOO")
	formatted := fmt.Sprintf("%+v", err)
	require.Contains(t, formatted, "FOO: test error")
	require.Contains(t, formatted, "github.com/Shopify/goose/bugsnag.TestFormatErrorClass")
	require.Contains(t, formatted, projectDir+"/bugsnag/with_class_test.go")
	require.Contains(t, formatted, "testing.tRunner")
	require.Contains(t, formatted, "src/testing/testing.go")
	require.Contains(t, formatted, "runtime.goexit")
	require.Contains(t, formatted, "src/runtime/asm_amd64.s")
}

func TestWrapf(t *testing.T) {
	err := Wrapf(errors.New("test error"), "other %v %v", "bar", "baz")
	require.Equal(t, "other bar baz", extractErrorClass(err))
}
