package sharedstore

import (
	"bytes"
	"encoding/gob"
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
		return nil, err
	}

	return decodeMemcacheItem(mItem)
}

func (w *memcacheClientWrapper) Set(key string, item *Item) error {
	mItem, err := encodeMemcacheItem(key, item)
	if err != nil {
		return err
	}
	return w.client.Set(mItem)
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
	return err
}

func (w *memcacheClientWrapper) Delete(key string) error {
	return w.client.Delete(key)
}
