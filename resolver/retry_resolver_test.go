package resolver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var temporaryError = &net.DNSError{Name: "foo", Err: "bar", IsTemporary: true}
var permanentError = &net.DNSError{Name: "foo", Err: "baz"}

func TestServFail(t *testing.T) {
	t.Skip("Development test only")

	for _, preferGo := range []bool{true, false} {
		t.Run(fmt.Sprintf("preferGo %t", preferGo), func(t *testing.T) {
			r := NewNetResolver(&net.Resolver{PreferGo: preferGo, StrictErrors: true})
			r = NewRetryResolver(r, []time.Duration{0}) // Retry once

			ctx := context.Background()
			_, err := r.LookupTXT(ctx, "camera-de-surveillance.com") // Misbehaving server, always returning SERVFAIL

			var dnsError *net.DNSError
			require.True(t, errors.As(err, &dnsError))
			require.False(t, dnsError.Temporary())
			require.Equal(t, dnsError.Err, "server misbehaving")
		})
	}
}

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
		call       func(ctx context.Context, t *testing.T, r Resolver, success bool) error
	}{
		"LookupHost": {
			callArgs:   []interface{}{mock.Anything, "foo"},
			returnArgs: []interface{}{[]string{"bar"}, nil},
			call: func(ctx context.Context, t *testing.T, r Resolver, success bool) error {
				addrs, err := r.LookupHost(ctx, "foo")
				if success {
					require.Equal(t, []string{"bar"}, addrs)
				}
				return err
			},
		},
		"LookupIPAddr": {
			callArgs:   []interface{}{mock.Anything, "foo"},
			returnArgs: []interface{}{[]net.IPAddr{{Zone: "bar"}}, nil},
			call: func(ctx context.Context, t *testing.T, r Resolver, success bool) error {
				records, err := r.LookupIPAddr(ctx, "foo")
				if success {
					require.Equal(t, []net.IPAddr{{Zone: "bar"}}, records)
				}
				return err
			},
		},
		"LookupPort": {
			callArgs:   []interface{}{mock.Anything, "a", "b"},
			returnArgs: []interface{}{123, nil},
			call: func(ctx context.Context, t *testing.T, r Resolver, success bool) error {
				port, err := r.LookupPort(ctx, "a", "b")
				if success {
					require.Equal(t, 123, port)
				}
				return err
			},
		},
		"LookupCNAME": {
			callArgs:   []interface{}{mock.Anything, "foo"},
			returnArgs: []interface{}{"bar", nil},
			call: func(ctx context.Context, t *testing.T, r Resolver, success bool) error {
				cname, err := r.LookupCNAME(ctx, "foo")
				if success {
					require.Equal(t, "bar", cname)
				}
				return err
			},
		},
		"LookupSRV": {
			callArgs:   []interface{}{mock.Anything, "a", "b", "c"},
			returnArgs: []interface{}{"bar", []*net.SRV{{Target: "bar"}}, nil},
			call: func(ctx context.Context, t *testing.T, r Resolver, success bool) error {
				cname, records, err := r.LookupSRV(ctx, "a", "b", "c")
				if success {
					require.Equal(t, "bar", cname)
					require.Equal(t, []*net.SRV{{Target: "bar"}}, records)
				}
				return err
			},
		},
		"LookupMX": {
			callArgs:   []interface{}{mock.Anything, "foo"},
			returnArgs: []interface{}{[]*net.MX{{Host: "bar"}}, nil},
			call: func(ctx context.Context, t *testing.T, r Resolver, success bool) error {
				records, err := r.LookupMX(ctx, "foo")
				if success {
					require.Equal(t, []*net.MX{{Host: "bar"}}, records)
				}
				return err
			},
		},
		"LookupNS": {
			callArgs:   []interface{}{mock.Anything, "foo"},
			returnArgs: []interface{}{[]*net.NS{{Host: "bar"}}, nil},
			call: func(ctx context.Context, t *testing.T, r Resolver, success bool) error {
				records, err := r.LookupNS(ctx, "foo")
				if success {
					require.Equal(t, []*net.NS{{Host: "bar"}}, records)
				}
				return err
			},
		},
		"LookupTXT": {
			callArgs:   []interface{}{mock.Anything, "foo"},
			returnArgs: []interface{}{[]string{"bar"}, nil},
			call: func(ctx context.Context, t *testing.T, r Resolver, success bool) error {
				records, err := r.LookupTXT(ctx, "foo")
				if success {
					require.Equal(t, []string{"bar"}, records)
				}
				return err
			},
		},
		"LookupAddr": {
			callArgs:   []interface{}{mock.Anything, "foo"},
			returnArgs: []interface{}{[]string{"bar"}, nil},
			call: func(ctx context.Context, t *testing.T, r Resolver, success bool) error {
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
					err := tt.call(ctx, t, r, false)
					require.EqualError(t, err, "lookup foo: baz")
				})
			})

			t.Run("retry servfail", func(t *testing.T) {
				withRetry(t, func(m *mockResolver, r Resolver) {
					m.On(method, tt.callArgs...).Return(makeErrorArgs(len(tt.returnArgs), &net.DNSError{Name: "foo", Err: "server misbehaving", IsTemporary: true})...).Times(3)
					err := tt.call(ctx, t, r, false)

					var dnsError *net.DNSError
					require.True(t, errors.As(err, &dnsError))
					require.False(t, dnsError.Temporary())
				})
			})

			t.Run("retry exhaustion", func(t *testing.T) {
				withRetry(t, func(m *mockResolver, r Resolver) {
					m.On(method, tt.callArgs...).Return(makeErrorArgs(len(tt.returnArgs), temporaryError)...).Times(3)
					err := tt.call(ctx, t, r, false)
					require.EqualError(t, err, "lookup foo: bar")
				})
			})

			t.Run("canceled", func(t *testing.T) {
				withRetry(t, func(m *mockResolver, r Resolver) {
					ctx, cancel := context.WithCancel(ctx)
					cancel()

					m.On(method, tt.callArgs...).Return(makeErrorArgs(len(tt.returnArgs), context.Canceled)...).Once()
					err := tt.call(ctx, t, r, false)
					require.EqualError(t, err, "context canceled")
				})
			})

			t.Run("deadline exceeded", func(t *testing.T) {
				withRetry(t, func(m *mockResolver, r Resolver) {
					m.On(method, tt.callArgs...).Return(makeErrorArgs(len(tt.returnArgs), context.DeadlineExceeded)...).Times(3)
					err := tt.call(ctx, t, r, false)
					require.EqualError(t, err, "context deadline exceeded")
				})
			})

			t.Run("success", func(t *testing.T) {
				withRetry(t, func(m *mockResolver, r Resolver) {
					m.On(method, tt.callArgs...).Return(makeErrorArgs(len(tt.returnArgs), temporaryError)...).Once()
					m.On(method, tt.callArgs...).Return(tt.returnArgs...).Once()
					err := tt.call(ctx, t, r, true)
					require.NoError(t, err)
				})
			})
		})
	}
}
