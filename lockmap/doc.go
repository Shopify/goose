// Package lockmap provides a thread-safe map of auto-expiring locks with promises to wait on.
// The underlying map will be periodically sweeped to remove expired promises.
// All promises returned by the map are guaranteed to be resolved roughly within its expiry.
// If a promise expires, is resolved manually, or is replaced, the channel will be closed.
package lockmap
