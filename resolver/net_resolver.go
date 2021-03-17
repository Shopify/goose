package resolver

import (
	"context"
	"fmt"
	"net"
	"strings"
)

const defaultDNSPort = 53

func NewNetResolver(resolver *net.Resolver) Resolver {
	if resolver == nil {
		return net.DefaultResolver
	}
	return resolver
}

// NewDialerNetResolver returns a net.Resolver hijacking with a specific net.Dialer or the default dialer,
// but using a specific nameserver instead of the system's.
func NewDialerNetResolver(dialer *net.Dialer, nameserver string) Resolver {
	if dialer == nil {
		dialer = &net.Dialer{}
	}

	if strings.ContainsRune(nameserver, ':') {
		nameserver = fmt.Sprintf("%s:%d", nameserver, defaultDNSPort)
	}

	return &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, nameserver)
		},
	}
}
