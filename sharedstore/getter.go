package sharedstore

import (
	"context"
	"time"

	"github.com/Shopify/go-cache/pkg"

	"github.com/Shopify/goose/lockmap"
)

// Getter can wait for its internal condition to be ready,
// such that it can return the desired data.
type Getter interface {
	Wait(ctx context.Context) (*cache.Item, error)
	WouldWait(ctx context.Context) bool
}

// resolvedGetter is essentially a noop, the data is already available.
type resolvedGetter struct {
	item *cache.Item
}

func (g *resolvedGetter) WouldWait(ctx context.Context) bool {
	return false
}

func (g *resolvedGetter) Wait(ctx context.Context) (*cache.Item, error) {
	return g.item, ctx.Err()
}

// promiseGetter waits on a condition to be signaled.
// Typically, it waits for another thread from the _same_ store instance to finish.
type promiseGetter struct {
	key     string
	store   Store
	promise lockmap.Promise
}

func (g *promiseGetter) WouldWait(ctx context.Context) bool {
	select {
	case <-g.promise:
		return false
	case <-ctx.Done():
		return false
	default:
		return true
	}
}

func (g *promiseGetter) Wait(ctx context.Context) (*cache.Item, error) {
	select {
	case <-g.promise:
		return g.store.getData(ctx, g.key)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

const pollingInterval = 100 * time.Millisecond

// pollGetter polls the store periodically until the key is unlocked.
// Typically, it waits for another thread on _another_ store instance to finish.
type pollGetter struct {
	key   string
	store Store
}

func (g *pollGetter) WouldWait(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
		// Checking this would require a call to the shared store, so assume it would wait.
		return true
	}
}

func (g *pollGetter) Wait(ctx context.Context) (*cache.Item, error) {
	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			locked, err := g.store.isLocked(ctx, g.key)
			if err != nil || !locked {
				// Broadcast even when there is an error so we unlock the threads
				// If the other threads can't find the value, they will simply try to build the item.
				g.store.broadcast(ctx, g.key)

				if err != nil {
					return nil, err
				}

				return g.store.getData(ctx, g.key)
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
