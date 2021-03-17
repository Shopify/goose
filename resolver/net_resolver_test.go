package resolver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewNetResolver(t *testing.T) {
	t.Skip("only meant to ran manually")

	r := NewNetResolver(nil)
	ips, err := r.LookupHost(context.Background(), "shopify.com")
	require.NoError(t, err)
	require.NotEmpty(t, ips)
}

func TestNewDialerNetResolver(t *testing.T) {
	t.Skip("only meant to ran manually")

	r := NewDialerNetResolver(nil, "8.8.8.8")
	ips, err := r.LookupHost(context.Background(), "shopify.com")
	require.NoError(t, err)
	require.NotEmpty(t, ips)
}
