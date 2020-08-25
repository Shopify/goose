package resolver

import (
	"net"
)

func NewNetResolver(resolver *net.Resolver) Resolver {
	if resolver == nil {
		return net.DefaultResolver
	}
	return resolver
}
