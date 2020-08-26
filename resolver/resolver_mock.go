package resolver

import (
	"context"
	"net"

	"github.com/stretchr/testify/mock"
)

type mockResolver struct {
	mock.Mock
}

var _ Resolver = (*mockResolver)(nil)

func NewMockResolver() *mockResolver {
	return &mockResolver{}
}

func (m *mockResolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	args := m.Called(ctx, host)
	if args.Error(1) != nil {
		return addrs, args.Error(1)
	}
	return args.Get(0).([]string), nil
}

func (m *mockResolver) LookupIPAddr(ctx context.Context, host string) (records []net.IPAddr, err error) {
	args := m.Called(ctx, host)
	if args.Error(1) != nil {
		return records, args.Error(1)
	}
	return args.Get(0).([]net.IPAddr), nil
}

func (m *mockResolver) LookupPort(ctx context.Context, network, service string) (port int, err error) {
	args := m.Called(ctx, network, service)
	if args.Error(1) != nil {
		return port, args.Error(1)
	}
	return args.Int(0), nil
}

func (m *mockResolver) LookupCNAME(ctx context.Context, host string) (cname string, err error) {
	args := m.Called(ctx, host)
	if args.Error(1) != nil {
		return cname, args.Error(1)
	}
	return args.String(0), nil
}

func (m *mockResolver) LookupSRV(ctx context.Context, service, proto, name string) (cname string, records []*net.SRV, err error) {
	args := m.Called(ctx, service, proto, name)
	if args.Error(2) != nil {
		return cname, records, args.Error(2)
	}
	return args.String(0), args.Get(1).([]*net.SRV), nil
}

func (m *mockResolver) LookupMX(ctx context.Context, name string) (records []*net.MX, err error) {
	args := m.Called(ctx, name)
	if args.Error(1) != nil {
		return records, args.Error(1)
	}
	return args.Get(0).([]*net.MX), nil
}

func (m *mockResolver) LookupNS(ctx context.Context, name string) (records []*net.NS, err error) {
	args := m.Called(ctx, name)
	if args.Error(1) != nil {
		return records, args.Error(1)
	}
	return args.Get(0).([]*net.NS), nil
}

func (m *mockResolver) LookupTXT(ctx context.Context, name string) (records []string, err error) {
	args := m.Called(ctx, name)
	if args.Error(1) != nil {
		return records, args.Error(1)
	}
	return args.Get(0).([]string), nil
}

func (m *mockResolver) LookupAddr(ctx context.Context, addr string) (names []string, err error) {
	args := m.Called(ctx, addr)
	if args.Error(1) != nil {
		return names, args.Error(1)
	}
	return args.Get(0).([]string), nil
}
