package resolver

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var temporaryError = &net.DNSError{Name: "foo", Err: "bar", IsTemporary: true}
var permanentError = &net.DNSError{Name: "foo", Err: "baz"}

func withRetry(t *testing.T, fn func(m *mockResolver, r Resolver)) {
	m := NewMockResolver()
	r := NewRetryResolver(m, []time.Duration{0, 0}) // Do not actually sleep in tests.

	fn(m, r)

	m.AssertExpectations(t)
}

func makeErrorArgs(length int, err error) []interface{} {
	args := make([]interface{}, length)
	args[len(args)-1] = err
	return args
}

func TestNewRetryLookup(t *testing.T) {
	ctx := context.Background()

	tests := map[string]struct {
		callArgs   []interface{}
		returnArgs []interface{}
		call       func(t *testing.T, r Resolver, success bool) error
	}{
		"LookupHost": {
			callArgs:   []interface{}{ctx, "foo"},
			returnArgs: []interface{}{[]string{"bar"}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				addrs, err := r.LookupHost(ctx, "foo")
				if success {
					require.Equal(t, []string{"bar"}, addrs)
				}
				return err
			},
		},
		"LookupIPAddr": {
			callArgs:   []interface{}{ctx, "foo"},
			returnArgs: []interface{}{[]net.IPAddr{{Zone: "bar"}}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				records, err := r.LookupIPAddr(ctx, "foo")
				if success {
					require.Equal(t, []net.IPAddr{{Zone: "bar"}}, records)
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
			returnArgs: []interface{}{"bar", []*net.SRV{{Target: "bar"}}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				cname, records, err := r.LookupSRV(ctx, "a", "b", "c")
				if success {
					require.Equal(t, "bar", cname)
					require.Equal(t, []*net.SRV{{Target: "bar"}}, records)
				}
				return err
			},
		},
		"LookupMX": {
			callArgs:   []interface{}{ctx, "foo"},
			returnArgs: []interface{}{[]*net.MX{{Host: "bar"}}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				records, err := r.LookupMX(ctx, "foo")
				if success {
					require.Equal(t, []*net.MX{{Host: "bar"}}, records)
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
			returnArgs: []interface{}{[]string{"bar"}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				records, err := r.LookupTXT(ctx, "foo")
				if success {
					require.Equal(t, []string{"bar"}, records)
				}
				return err
			},
		},
		"LookupAddr": {
			callArgs:   []interface{}{ctx, "foo"},
			returnArgs: []interface{}{[]string{"bar"}, nil},
			call: func(t *testing.T, r Resolver, success bool) error {
				names, err := r.LookupAddr(ctx, "foo")
				if success {
					require.Equal(t, []string{"bar"}, names)
				}
				return err
			},
		},
	}

	for method, tt := range tests {
		t.Run(method, func(t *testing.T) {
			t.Run("retry temporary", func(t *testing.T) {
				withRetry(t, func(m *mockResolver, r Resolver) {
					m.On(method, tt.callArgs...).Return(makeErrorArgs(len(tt.returnArgs), temporaryError)...).Once()
					m.On(method, tt.callArgs...).Return(makeErrorArgs(len(tt.returnArgs), permanentError)...).Once()
					err := tt.call(t, r, false)
					require.EqualError(t, err, "lookup foo: baz")
				})
			})

			t.Run("retry exhaustion", func(t *testing.T) {
				withRetry(t, func(m *mockResolver, r Resolver) {
					m.On(method, tt.callArgs...).Return(makeErrorArgs(len(tt.returnArgs), temporaryError)...).Times(3)
					err := tt.call(t, r, false)
					require.EqualError(t, err, "lookup foo: bar")
				})
			})

			t.Run("success", func(t *testing.T) {
				withRetry(t, func(m *mockResolver, r Resolver) {
					m.On(method, tt.callArgs...).Return(makeErrorArgs(len(tt.returnArgs), temporaryError)...).Once()
					m.On(method, tt.callArgs...).Return(tt.returnArgs...).Once()
					err := tt.call(t, r, true)
					require.NoError(t, err)
				})
			})
		})
	}
}
