package httperrors

import (
	"errors"
	stderrors "errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHttpErrorBuilder(t *testing.T) {
	orig := stderrors.New("test")
	he := New(orig).WithStatus(http.StatusInternalServerError).WithMsg("Server dropped the ball")

	require.NotNil(t, he)

	unwrapped := stderrors.Unwrap(he)
	require.NotNil(t, unwrapped)
	require.True(t, stderrors.Is(unwrapped, errors.New("test")))
}
