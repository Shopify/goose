package errors

import (
	stderrors "errors"
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

func joinedError() error {
	return stderrors.Join(New("first"), New("second"))
}

func nestedJoinedError() error {
	return Wrap(stderrors.Join(New("second"), New("third")), "first")
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
		{
			test:       "baseError-stdJoinedError",
			wrappedErr: joinedError,
			stackLen:   3,
		},
		{
			test:       "baseError-stdNestedJoinedError",
			wrappedErr: nestedJoinedError,
			stackLen:   4,
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
