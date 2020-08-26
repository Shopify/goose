package resolver

import (
	"context"
	"net"
	"testing"

	"github.com/Shopify/goose/random"

	"github.com/stretchr/testify/require"
)

func withShuffle(t *testing.T, fn func(m *mockResolver, r Resolver)) {
	m := NewMockResolver()
	r := NewShuffleResolver(m).(*shuffleResolver)
	r.rand = random.NewDummy()

	fn(m, r)

	m.AssertExpectations(t)
}

func TestNewShuffleResolver(t *testing.T) {
	ctx := context.Background()

	tests := map[string]struct {
		callArgs   []interface{}
		returnArgs []interface{}
		call       func(t *testing.T, r Resolver, success bool) error
	}{
		"LookupHost": {
			callArgs:   []interface{}{ctx, "foo"},
			returnArgs: []interface{}{[]string{"bar", "baz", "qux"}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				addrs, err := r.LookupHost(ctx, "foo")
				if success {
					require.Equal(t, []string{"baz", "bar", "qux"}, addrs)
				}
				return err
			},
		},
		"LookupIPAddr": {
			callArgs:   []interface{}{ctx, "foo"},
			returnArgs: []interface{}{[]net.IPAddr{{Zone: "bar"}, {Zone: "baz"}, {Zone: "qux"}}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				records, err := r.LookupIPAddr(ctx, "foo")
				if success {
					require.Equal(t, []net.IPAddr{{Zone: "baz"}, {Zone: "bar"}, {Zone: "qux"}}, records)
				}
				return err
			},
		},
		"LookupPort": {
			callArgs:   []interface{}{ctx, "a", "b"},
			returnArgs: []interface{}{123, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				port, err := r.LookupPort(ctx, "a", "b")
				if success {
					require.Equal(t, 123, port)
				}
				return err
			},
		},
		"LookupCNAME": {
			callArgs:   []interface{}{ctx, "foo"},
			returnArgs: []interface{}{"bar", nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				cname, err := r.LookupCNAME(ctx, "foo")
				if success {
					require.Equal(t, "bar", cname)
				}
				return err
			},
		},
		"LookupSRV": {
			callArgs:   []interface{}{ctx, "a", "b", "c"},
			returnArgs: []interface{}{"bar", []*net.SRV{{Target: "bar"}, {Target: "baz"}, {Target: "qux"}}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				cname, records, err := r.LookupSRV(ctx, "a", "b", "c")
				if success {
					require.Equal(t, "bar", cname)
					require.Equal(t, []*net.SRV{{Target: "baz"}, {Target: "bar"}, {Target: "qux"}}, records)
				}
				return err
			},
		},
		"LookupMX": {
			callArgs:   []interface{}{ctx, "foo"},
			returnArgs: []interface{}{[]*net.MX{{Host: "bar"}, {Host: "baz"}, {Host: "qux"}}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				records, err := r.LookupMX(ctx, "foo")
				if success {
					require.Equal(t, []*net.MX{{Host: "baz"}, {Host: "bar"}, {Host: "qux"}}, records)
				}
				return err
			},
		},
		"LookupNS": {
			callArgs:   []interface{}{ctx, "foo"},
			returnArgs: []interface{}{[]*net.NS{{Host: "bar"}}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				records, err := r.LookupNS(ctx, "foo")
				if success {
					require.Equal(t, []*net.NS{{Host: "bar"}}, records)
				}
				return err
			},
		},
		"LookupTXT": {
			callArgs:   []interface{}{ctx, "foo"},
			returnArgs: []interface{}{[]string{"bar", "baz", "qux"}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				records, err := r.LookupTXT(ctx, "foo")
				if success {
					require.Equal(t, []string{"baz", "bar", "qux"}, records)
				}
				return err
			},
		},
		"LookupAddr": {
			callArgs:   []interface{}{ctx, "foo"},
			returnArgs: []interface{}{[]string{"bar", "baz", "qux"}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				names, err := r.LookupAddr(ctx, "foo")
				if success {
					require.Equal(t, []string{"baz", "bar", "qux"}, names)
				}
				return err
			},
		},
	}

	for method, tt := range tests {
		t.Run(method, func(t *testing.T) {
			t.Run("error", func(t *testing.T) {
				withShuffle(t, func(m *mockResolver, r Resolver) {
					m.On(method, tt.callArgs...).Return(makeErrorArgs(len(tt.returnArgs), temporaryError)...).Once()
					err := tt.call(t, r, false)
					require.EqualError(t, err, "lookup foo: bar")
				})
			})

			t.Run("success", func(t *testing.T) {
				withShuffle(t, func(m *mockResolver, r Resolver) {
					m.On(method, tt.callArgs...).Return(tt.returnArgs...).Once()
					err := tt.call(t, r, true)
					require.NoError(t, err)
				})
			})
		})
	}
}
