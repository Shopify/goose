package sharedstore

import (
	"context"
	"net"
	"time"

	cache "github.com/Shopify/go-cache/pkg"
	"github.com/pkg/errors"
	"gopkg.in/tomb.v2"

	"github.com/Shopify/goose/lockmap"
	"github.com/Shopify/goose/logger"
)

var log = logger.New("sharedstore")

// Store is able to retrieve data from a Client, or take a lock so the data can be set later
type Store interface {
	// GetOrLock returns a Getter, which is able to retrieve the data and/or a Setter,
	// which can be invoked to set the data and release the locks.
	GetOrLock(ctx context.Context, key string) (Getter, Setter)
	getData(ctx context.Context, key string) (*cache.Item, error)
	setData(ctx context.Context, key string, data interface{}, ttl time.Duration) (*cache.Item, error)

	isLocked(ctx context.Context, key string) (bool, error)
	unlock(ctx context.Context, key string) error

	// broadcast notifies all pending threads of this instance that the key is unlocked and ready for retrieval
	broadcast(ctx context.Context, key string)

	Run() error
	Tomb() *tomb.Tomb
}

func New(client cache.Client, lockExpiry time.Duration) Store {
	return &store{
		client:     client,
		lockExpiry: lockExpiry,
		lockMap:    lockmap.New(lockExpiry/10, &tomb.Tomb{}),
	}
}

type store struct {
	client     cache.Client
	lockExpiry time.Duration

	// All threads waiting on the same key will subscribe to the condition
	// When adding to the map, make sure to remove the condition when done.
	// See the goroutine in GetOrLock that schedules a removal after lockExpiry.
	lockMap lockmap.LockMap
}

func (s *store) Run() error {
	return s.lockMap.Run()
}

func (s *store) Tomb() *tomb.Tomb {
	return s.lockMap.Tomb()
}

func (s *store) GetOrLock(ctx context.Context, key string) (Getter, Setter) {
	item, err := s.getData(ctx, key)
	if err != nil {
		if isTemporaryError(err) {
			log(ctx, err).Warn("temporary client error")
		} else {
			log(ctx, err).Error("unexpected client error")
		}
	} else if item != nil {
		return &resolvedGetter{
			item: item,
		}, nil
	}

	promise, gotLock := s.lockMap.WaitOrLock(key, s.lockExpiry)

	// This is to populate the data, it may not be needed by the caller if the subscribe succeeds
	setter := &setter{
		store: s,
		key:   key,
	}

	if !gotLock {
		// A thread on this instance has the lock
		return &promiseGetter{
			promise: promise,
			key:     key,
			store:   s,
		}, setter
	}

	if ok, err := s.lock(ctx, key); err != nil {
		if isTemporaryError(err) {
			log(ctx, err).Warn("temporarily unable to lock item")
		} else {
			log(ctx, err).Error("unable to lock item")
		}
		// We don't have the memcache lock, but we still have the local lock,
		// which mitigates some of the concurrency.
		// Proceed with the setter, to make sure threads get unlocked.
		return nil, setter
	} else if !ok {
		// A thread on another instance has the lock
		return &pollGetter{
			key:   key,
			store: s,
		}, setter
	}

	return nil, setter
}

func (s *store) getData(ctx context.Context, key string) (*cache.Item, error) {
	log(ctx, nil).WithField("key", key).Debug("retrieving item")

	return s.client.Get(key)
}

func (s *store) setData(ctx context.Context, key string, data interface{}, ttl time.Duration) (*cache.Item, error) {
	log(ctx, nil).WithField("key", key).Debug("setting item")

	item := cache.Item{
		Data:       data,
		Expiration: time.Now().Add(ttl),
	}

	if err := s.client.Set(key, &item); err != nil {
		return nil, errors.Wrap(err, "unable to set item")
	}

	return &item, nil
}

func (s *store) lock(ctx context.Context, key string) (bool, error) {
	log(ctx, nil).WithField("key", key).Debug("locking item")

	err := s.client.Add(key+".lock", &cache.Item{
		Expiration: time.Now().Add(s.lockExpiry),
	})
	if err == nil {
		return true, nil
	}

	if err == cache.ErrNotStored {
		log(ctx, err).Info("lock belongs to another instance")
		err = nil
	}

	return false, err
}

func (s *store) isLocked(ctx context.Context, key string) (bool, error) {
	log(ctx, nil).WithField("key", key).Debug("checking item lock")

	item, err := s.client.Get(key + ".lock")
	locked := item != nil
	return locked, err
}

func (s *store) unlock(ctx context.Context, key string) error {
	log(ctx, nil).WithField("key", key).Debug("unlocking item")

	return s.client.Delete(key + ".lock")
}

func (s *store) broadcast(ctx context.Context, key string) {
	log(ctx, nil).WithField("key", key).Info("broadcasting to other item threads")

	s.lockMap.Release(key)
}

func isTemporaryError(err error) bool {
	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary()
	}
	return false
}
