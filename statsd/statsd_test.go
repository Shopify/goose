package statsd

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestBackend(t *testing.T) {
	dd, err := NewDatadogBackend("localhost:8125", "catwalk", []string{"global:tag"})
	require.NoError(t, err)

	tests := []struct {
		impl string
		exp  Backend
		err  error
	}{
		{impl: "datadog", exp: dd},
		{impl: "null", exp: NewNullBackend()},
		{impl: "log", exp: NewLogBackend("catwalk", []string{})},
		{err: ErrUnknownBackend},
		{impl: "WHAT", err: ErrUnknownBackend},
	}

	for _, test := range tests {
		currentBackend = nil
		b, err := NewBackend(test.impl, "localhost:8125", "catwalk", "global:tag")
		if test.err != nil {
			fmt.Printf("testing for: %s\n", test.impl)
			require.Nil(t, currentBackend)
			require.Equal(t, test.err, errors.Cause(err))
			continue
		}

		require.IsType(t, test.exp, b)
		require.NoError(t, err)
	}
}
