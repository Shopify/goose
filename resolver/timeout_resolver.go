package resolver

import (
	"context"
	"net"
	"time"
)

type timeoutResolver struct {
	lookup  Resolver
	timeout time.Duration
}

func NewTimeoutResolver(lookup Resolver, timeout time.Duration) Resolver {
	return &timeoutResolver{
		lookup:  lookup,
		timeout: timeout,
	}
}

func (r *timeoutResolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.lookup.LookupHost(ctx, host)
}

func (r *timeoutResolver) LookupIPAddr(ctx context.Context, host string) (addrs []net.IPAddr, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.lookup.LookupIPAddr(ctx, host)
}

func (r *timeoutResolver) LookupPort(ctx context.Context, network, service string) (port int, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.lookup.LookupPort(ctx, network, service)
}

func (r *timeoutResolver) LookupCNAME(ctx context.Context, host string) (cname string, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.lookup.LookupCNAME(ctx, host)
}

func (r *timeoutResolver) LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.lookup.LookupSRV(ctx, service, proto, name)
}

func (r *timeoutResolver) LookupMX(ctx context.Context, name string) (records []*net.MX, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.lookup.LookupMX(ctx, name)
}

func (r *timeoutResolver) LookupNS(ctx context.Context, name string) (records []*net.NS, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.lookup.LookupNS(ctx, name)
}

func (r *timeoutResolver) LookupTXT(ctx context.Context, name string) (records []string, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.lookup.LookupTXT(ctx, name)
}

func (r *timeoutResolver) LookupAddr(ctx context.Context, addr string) (names []string, err error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	return r.lookup.LookupAddr(ctx, addr)
}
