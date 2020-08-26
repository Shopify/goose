// Package resolver provides an interface and wrappers to the net.Resolver
package resolver

// Retry, defaulting to 3 tries will exponential backoff
// Shuffle, to randomize the outputs because some systems to not randomize theirs.
// Mock, for testing.
//
// resolver.New() will return a resolver wrapping the default net.Resolver in a shuffle and retry logic.
