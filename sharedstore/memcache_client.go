package sharedstore

import (
	"bytes"
	"encoding/gob"
	"net"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/pkg/errors"
)

// memcacheClient is essentially a *memcached.Client, but this allows it to be mocked
type memcacheClient interface {
	Get(key string) (*memcache.Item, error)
	Set(item *memcache.Item) error
	Add(item *memcache.Item) error
	Delete(key string) error
}

func NewMemcacheClient(c memcacheClient) Client {
	return &memcacheClientWrapper{client: c}
}

func decodeMemcacheItem(mItem *memcache.Item) (*Item, error) {
	dec := gob.NewDecoder(bytes.NewReader(mItem.Value))
	var item Item
	err := dec.Decode(&item)
	return &item, errors.Wrap(err, "unable to decode item")
}

func encodeMemcacheItem(key string, item *Item) (*memcache.Item, error) {
	encoded := &bytes.Buffer{}
	enc := gob.NewEncoder(encoded)
	if err := enc.Encode(*item); err != nil {
		return nil, errors.Wrap(err, "unable to encode item")
	}

	return &memcache.Item{
		Value:      encoded.Bytes(),
		Expiration: int32(time.Until(item.Expiration).Seconds()),
		Key:        key,
	}, nil
}

type memcacheClientWrapper struct {
	client memcacheClient
}

func (w *memcacheClientWrapper) Get(key string) (*Item, error) {
	mItem, err := w.client.Get(key)
	if err != nil {
		// Abstract the memcache-specific error
		if err == memcache.ErrCacheMiss {
			err = nil
		}
		return nil, coalesceTimeoutError(err)
	}

	return decodeMemcacheItem(mItem)
}

func (w *memcacheClientWrapper) Set(key string, item *Item) error {
	mItem, err := encodeMemcacheItem(key, item)
	if err != nil {
		return err
	}
	return coalesceTimeoutError(w.client.Set(mItem))
}

func (w *memcacheClientWrapper) Add(key string, item *Item) error {
	mItem, err := encodeMemcacheItem(key, item)
	if err != nil {
		return err
	}
	err = w.client.Set(mItem)

	if err == memcache.ErrNotStored {
		// Abstract the memcache-specific error
		return ErrNotStored
	}
	return coalesceTimeoutError(err)
}

func (w *memcacheClientWrapper) Delete(key string) error {
	err := w.client.Delete(key)
	if err == memcache.ErrCacheMiss {
		// Deleting a missing entry is not an actual issue.
		return nil
	}
	return coalesceTimeoutError(err)
}

type connectTimeoutError struct{}

func (connectTimeoutError) Error() string   { return "memcache: connect timeout" }
func (connectTimeoutError) Timeout() bool   { return true }
func (connectTimeoutError) Temporary() bool { return true }

func coalesceTimeoutError(err error) error {
	// For some reason, gomemcache decided to replace the standard net.Error.
	// Coalesce into a generic net.Error so that client don't have to deal with memcache-specific errors.
	switch typed := err.(type) {
	case *memcache.ConnectTimeoutError:
		return &net.OpError{
			Err:  &connectTimeoutError{},
			Addr: typed.Addr,
			Net:  typed.Addr.Network(),
			Op:   "connect",
		}
	default:
		// This also work if err is nil
		return err
	}
}
