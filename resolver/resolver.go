package resolver

import (
	"context"
	"errors"
	"net"
)

var ErrNotFound = errors.New("host not found")

type Resolver interface {
	LookupHost(ctx context.Context, host string) (addrs []string, err error)
	LookupIPAddr(ctx context.Context, host string) ([]net.IPAddr, error)
	LookupPort(ctx context.Context, network, service string) (port int, err error)
	LookupCNAME(ctx context.Context, host string) (cname string, err error)
	LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error)
	LookupMX(ctx context.Context, name string) ([]*net.MX, error)
	LookupNS(ctx context.Context, name string) ([]*net.NS, error)
	LookupTXT(ctx context.Context, name string) ([]string, error)
	LookupAddr(ctx context.Context, addr string) (names []string, err error)
}

func New() (r Resolver) {
	r = NewNetResolver(nil)
	r = NewRetryResolver(r, nil)
	r = NewShuffleResolver(r)
	return r
}

func IsNotFound(err error) bool {
	var dnsError *net.DNSError
	return errors.As(err, &dnsError) && dnsError.IsNotFound
}
