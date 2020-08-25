package resolver

import (
	"context"
	"math/rand"
	"net"

	"github.com/Shopify/goose/random"
)

type shuffleResolver struct {
	lookup Resolver
	rand   *rand.Rand
}

func NewShuffleResolver(lookup Resolver) Resolver {
	return &shuffleResolver{
		lookup: lookup,
		rand:   random.New(),
	}
}

func (r *shuffleResolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	addrs, err = r.lookup.LookupHost(ctx, host)
	r.rand.Shuffle(len(addrs), func(i, j int) {
		addrs[i], addrs[j] = addrs[j], addrs[i]
	})
	return addrs, err
}

func (r *shuffleResolver) LookupIPAddr(ctx context.Context, host string) (addrs []net.IPAddr, err error) {
	addrs, err = r.lookup.LookupIPAddr(ctx, host)
	r.rand.Shuffle(len(addrs), func(i, j int) {
		addrs[i], addrs[j] = addrs[j], addrs[i]
	})
	return addrs, err
}

func (r *shuffleResolver) LookupPort(ctx context.Context, network, service string) (port int, err error) {
	return r.lookup.LookupPort(ctx, network, service)
}

func (r *shuffleResolver) LookupCNAME(ctx context.Context, host string) (cname string, err error) {
	return r.lookup.LookupCNAME(ctx, host)
}

func (r *shuffleResolver) LookupSRV(ctx context.Context, service, proto, name string) (cname string, addrs []*net.SRV, err error) {
	cname, addrs, err = r.lookup.LookupSRV(ctx, service, proto, name)
	r.rand.Shuffle(len(addrs), func(i, j int) {
		addrs[i], addrs[j] = addrs[j], addrs[i]
	})
	return cname, addrs, err
}

func (r *shuffleResolver) LookupMX(ctx context.Context, name string) (records []*net.MX, err error) {
	records, err = r.lookup.LookupMX(ctx, name)
	r.rand.Shuffle(len(records), func(i, j int) {
		records[i], records[j] = records[j], records[i]
	})
	return records, err
}

func (r *shuffleResolver) LookupNS(ctx context.Context, name string) (records []*net.NS, err error) {
	records, err = r.lookup.LookupNS(ctx, name)
	r.rand.Shuffle(len(records), func(i, j int) {
		records[i], records[j] = records[j], records[i]
	})
	return records, err
}

func (r *shuffleResolver) LookupTXT(ctx context.Context, name string) (records []string, err error) {
	records, err = r.lookup.LookupTXT(ctx, name)
	r.rand.Shuffle(len(records), func(i, j int) {
		records[i], records[j] = records[j], records[i]
	})
	return records, err
}

func (r *shuffleResolver) LookupAddr(ctx context.Context, addr string) (names []string, err error) {
	names, err = r.lookup.LookupAddr(ctx, addr)
	r.rand.Shuffle(len(names), func(i, j int) {
		names[i], names[j] = names[j], names[i]
	})
	return names, err
}
