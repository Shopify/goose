package httperrors

import (
	stderrors "errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHttpErrorBuilder(t *testing.T) {
	orig := stderrors.New("test")
	he := New(orig).WithStatus(http.StatusInternalServerError).WithMsg("Server dropped the ball")

	require.NotNil(t, he)

	unwrapped := stderrors.Unwrap(he)
	require.NotNil(t, unwrapped)
	require.True(t, stderrors.Is(unwrapped, orig))
}

func TestResponseWriter(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := stderrors.New("test")
		wrapped := New(err).WithStatus(http.StatusInternalServerError).WithMsg("something went wrong")
		wrapped.Write(w)
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}

	require.Equal(t, http.StatusInternalServerError, res.StatusCode)
	require.Equal(t, "500 Internal Server Error", res.Status)
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, "something went wrong", string(body))
}
