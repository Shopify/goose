package resolver

import (
	"context"
	"errors"
	"net"
	"time"
)

// Perform one immediate retry to limit the introduction of delays during DNS resolutions
var defaultRetryBackoffs = []time.Duration{0, 500 * time.Millisecond, 2 * time.Second}

type retryResolver struct {
	resolver Resolver
	backoffs []time.Duration
}

func NewRetryResolver(resolver Resolver, backoffs []time.Duration) Resolver {
	if backoffs == nil {
		backoffs = defaultRetryBackoffs
	}
	return &retryResolver{
		backoffs: backoffs,
		resolver: resolver,
	}
}

func shouldRetry(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var dnsError *net.DNSError
	return errors.As(err, &dnsError) && dnsError.Temporary()
}

func (r *retryResolver) retry(ctx context.Context, fn func() error) (err error) {
	err = fn()

	for i := 0; i < len(r.backoffs) && shouldRetry(err); i++ {
		select {
		case <-ctx.Done():
			// If the parent context is done, retrying wouldn't help, no need to check if it is a context.DeadlineExceeded.
			return ctx.Err()
		case <-time.After(r.backoffs[i]):
			err = fn()
		}
	}

	var dnsError *net.DNSError
	if errors.As(err, &dnsError) && dnsError.Err == "server misbehaving" {
		// Certain servers misbehave and will very likely continue to do so.
		// If the query was attempted several times and still results in a SERVFAIL, consider it permanent.

		// Note that this code doesn't distinguish the edge case where the previous errors were a different temporary error.

		// Modify the DNSError, regardless of whether it's wrapped.
		// Since "err" holds a pointer to the DNSError, the returned error will contain the modified DNSError.
		dnsError.IsTemporary = false
	}

	return err
}

func (r *retryResolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	err = r.retry(ctx, func() (err error) {
		addrs, err = r.resolver.LookupHost(ctx, host)
		return err
	})
	return addrs, err
}

func (r *retryResolver) LookupIPAddr(ctx context.Context, host string) (records []net.IPAddr, err error) {
	err = r.retry(ctx, func() (err error) {
		records, err = r.resolver.LookupIPAddr(ctx, host)
		return err
	})
	return records, err
}

func (r *retryResolver) LookupPort(ctx context.Context, network, service string) (port int, err error) {
	err = r.retry(ctx, func() (err error) {
		port, err = r.resolver.LookupPort(ctx, network, service)
		return err
	})
	return port, err
}

func (r *retryResolver) LookupCNAME(ctx context.Context, host string) (cname string, err error) {
	err = r.retry(ctx, func() (err error) {
		cname, err = r.resolver.LookupCNAME(ctx, host)
		return err
	})
	return cname, err
}

func (r *retryResolver) LookupSRV(ctx context.Context, service, proto, name string) (cname string, records []*net.SRV, err error) {
	err = r.retry(ctx, func() (err error) {
		cname, records, err = r.resolver.LookupSRV(ctx, service, proto, name)
		return err
	})
	return cname, records, err
}

func (r *retryResolver) LookupMX(ctx context.Context, name string) (records []*net.MX, err error) {
	err = r.retry(ctx, func() (err error) {
		records, err = r.resolver.LookupMX(ctx, name)
		return err
	})
	return records, err
}

func (r *retryResolver) LookupNS(ctx context.Context, name string) (records []*net.NS, err error) {
	err = r.retry(ctx, func() (err error) {
		records, err = r.resolver.LookupNS(ctx, name)
		return err
	})
	return records, err
}

func (r *retryResolver) LookupTXT(ctx context.Context, name string) (records []string, err error) {
	err = r.retry(ctx, func() (err error) {
		records, err = r.resolver.LookupTXT(ctx, name)
		return err
	})
	return records, err
}

func (r *retryResolver) LookupAddr(ctx context.Context, addr string) (names []string, err error) {
	err = r.retry(ctx, func() (err error) {
		names, err = r.resolver.LookupAddr(ctx, addr)
		return err
	})
	return names, err
}
