package cursor

import (
	"context"
	"errors"
	"fmt"

	"github.com/Shopify/go-cache/v2"
)

const cachePrefix = "task_cursors"

type Cursor interface {
	Current(ctx context.Context) (int, error)
	Increment(ctx context.Context) (int, error)
	Set(ctx context.Context, value int) error
}

type cursor struct {
	name  string
	cache cache.Client
}

func NewCursor(name string, cache cache.Client) Cursor {
	return &cursor{
		name:  fmt.Sprintf("%s_%s", cachePrefix, name),
		cache: cache,
	}
}

func (c *cursor) Current(ctx context.Context) (int, error) {
	var val uint64
	err := c.cache.Get(ctx, c.name, &val)
	if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		return 0, err
	}
	return int(val), nil
}

func (c *cursor) Increment(ctx context.Context) (int, error) {
	val, err := c.cache.Increment(ctx, c.name, uint64(1))
	return int(val), err
}

func (c *cursor) Set(ctx context.Context, value int) error {
	return c.cache.Set(ctx, c.name, uint64(value), cache.NeverExpire)
}
