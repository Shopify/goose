package errors

import (
	"testing"

	bugsnagerrors "github.com/bugsnag/bugsnag-go/v2/errors"
	pkgErrors "github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func rawStdError() error {
	return New("")
}

func nestedPkgError() error {
	return pkgErrors.Wrap(New(""), "")
}

func nestedBaseError() error {
	return Wrap(New(""), "")
}

func Test_baseError_Callers(t *testing.T) {
	tests := []struct {
		test       string
		wrappedErr func() error
		stackLen   int
	}{
		{
			test:       "baseError-stdError",
			wrappedErr: rawStdError,
			stackLen:   3, // asm_amd64.s - testing.go - Wrap(tt.wrappedErr(), "")
		},
		{
			test:       "baseError-baseError-stdError",
			wrappedErr: nestedBaseError,
			stackLen:   4, // asm_amd64.s - testing.go - Wrap(tt.wrappedErr(), "") - nestedBaseError
		},
		{
			test:       "baseError-pkgError-stdError",
			wrappedErr: nestedPkgError,
			stackLen:   4, // asm_amd64.s - testing.go - Wrap(tt.wrappedErr(), "") - nestedPkgError
		},
	}

	for _, tt := range tests {
		t.Run(tt.test, func(t *testing.T) {
			err := Wrap(tt.wrappedErr(), "")

			stack := err.(bugsnagerrors.ErrorWithCallers).Callers() //nolint:errorlint
			require.Len(t, stack, tt.stackLen, formatStack(stack))
		})
	}
}
