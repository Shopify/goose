// Package cond provides an implementation very similar to sync.Cond,
// but with support for cancellation due to timeouts, dead tombs, or done contexts.
// However, one major difference is that Signal() and Broadcast() both require the lock
package cond
