package resolver

import (
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		isNotFound bool
	}{
		{"nil", nil, false},
		{"foo", errors.New("foo"), false},
		{"dns error", &net.DNSError{}, false},
		{"dns not found error", &net.DNSError{IsNotFound: true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.isNotFound, IsNotFound(tt.err))
		})
	}
}
