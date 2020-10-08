package resolver

import (
	"context"
	"net"

	"github.com/Shopify/goose/statsd"
)

type metricResolver struct {
	resolver Resolver
	timer    *statsd.Timer
	tags     statsd.Tags
}

func NewWithMetrics(resolver Resolver, timer *statsd.Timer, tags statsd.Tags) Resolver {
	return &metricResolver{
		resolver: resolver,
		timer:    timer,
		tags:     tags,
	}
}

func (r *metricResolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	defer r.timer.StartTimer(ctx, statsd.Tags{"method": "LookupHost"}, r.tags).SuccessFinish(&err)
	return r.resolver.LookupHost(ctx, host)
}

func (r *metricResolver) LookupIPAddr(ctx context.Context, host string) (records []net.IPAddr, err error) {
	defer r.timer.StartTimer(ctx, statsd.Tags{"method": "LookupIPAddr"}, r.tags).SuccessFinish(&err)
	return r.resolver.LookupIPAddr(ctx, host)
}

func (r *metricResolver) LookupPort(ctx context.Context, network, service string) (port int, err error) {
	defer r.timer.StartTimer(ctx, statsd.Tags{"method": "LookupPort"}, r.tags).SuccessFinish(&err)
	return r.resolver.LookupPort(ctx, network, service)
}

func (r *metricResolver) LookupCNAME(ctx context.Context, host string) (cname string, err error) {
	defer r.timer.StartTimer(ctx, statsd.Tags{"method": "LookupCNAME"}, r.tags).SuccessFinish(&err)
	return r.resolver.LookupCNAME(ctx, host)
}

func (r *metricResolver) LookupSRV(ctx context.Context, service, proto, name string) (cname string, records []*net.SRV, err error) {
	defer r.timer.StartTimer(ctx, statsd.Tags{"method": "LookupSRV"}, r.tags).SuccessFinish(&err)
	return r.resolver.LookupSRV(ctx, service, proto, name)
}

func (r *metricResolver) LookupMX(ctx context.Context, name string) (records []*net.MX, err error) {
	defer r.timer.StartTimer(ctx, statsd.Tags{"method": "LookupMX"}, r.tags).SuccessFinish(&err)
	return r.resolver.LookupMX(ctx, name)
}

func (r *metricResolver) LookupNS(ctx context.Context, name string) (records []*net.NS, err error) {
	defer r.timer.StartTimer(ctx, statsd.Tags{"method": "LookupNS"}, r.tags).SuccessFinish(&err)
	return r.resolver.LookupNS(ctx, name)
}

func (r *metricResolver) LookupTXT(ctx context.Context, name string) (records []string, err error) {
	defer r.timer.StartTimer(ctx, statsd.Tags{"method": "LookupTXT"}, r.tags).SuccessFinish(&err)
	return r.resolver.LookupTXT(ctx, name)
}

func (r *metricResolver) LookupAddr(ctx context.Context, addr string) (names []string, err error) {
	defer r.timer.StartTimer(ctx, statsd.Tags{"method": "LookupAddr"}, r.tags).SuccessFinish(&err)
	return r.resolver.LookupAddr(ctx, addr)
}
