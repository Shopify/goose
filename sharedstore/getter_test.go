package sharedstore

import (
	"context"
	"testing"
	"time"

	cache "github.com/Shopify/go-cache/pkg"
	"github.com/stretchr/testify/assert"
)

func Test_promiseGetter(t *testing.T) {
	client := cache.NewMemoryClient()
	ctx := context.Background()
	store := New(client, time.Second)
	store.Tomb().Go(store.Run)

	sub, prov := store.GetOrLock(ctx, "foo")

	assert.Nil(t, sub, "should not have a getter")
	assert.IsType(t, &setter{}, prov, "should have a setter")

	sub2, prov2 := store.GetOrLock(ctx, "foo")
	assert.IsType(t, &promiseGetter{}, sub2, "should have a getter")
	assert.True(t, sub2.WouldWait(ctx))
	assert.IsType(t, &setter{}, prov2, "should have a setter")

	done := make(chan *cache.Item)
	go func() {
		item, err := sub2.Wait(ctx)
		assert.NoError(t, err)
		done <- item
	}()
	<-time.After(100 * time.Millisecond)

	err := prov.Done(ctx, "bar", 10*time.Second)
	assert.NoError(t, err)

	select {
	case item := <-done:
		assert.Equal(t, "bar", item.Data.(string))
		assert.False(t, sub2.WouldWait(ctx), "should not wait after done")
	case <-time.After(1 * time.Second):
		assert.Fail(t, "took too long to complete")
	}
}

func Test_pollGetter(t *testing.T) {
	client := cache.NewMemoryClient()
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	store := New(client, time.Second)
	store.Tomb().Go(store.Run)

	sub, prov := store.GetOrLock(ctx, "foo")

	assert.Nil(t, sub, "should not have a getter")
	assert.IsType(t, &setter{}, prov, "should have a setter")

	// New store to mimic a new instance
	store2 := New(client, time.Second)
	store2.Tomb().Go(store.Run)

	sub2, prov2 := store2.GetOrLock(ctx, "foo")
	assert.IsType(t, &pollGetter{}, sub2, "should have a getter")
	assert.True(t, sub2.WouldWait(ctx))
	assert.IsType(t, &setter{}, prov2, "should have a setter")

	done := make(chan *cache.Item)
	go func() {
		item, err := sub2.Wait(ctx)
		assert.NoError(t, err)
		done <- item
	}()
	<-time.After(100 * time.Millisecond)

	err := prov.Done(ctx, "bar", 10*time.Second)
	assert.NoError(t, err)

	select {
	case item := <-done:
		assert.Equal(t, "bar", item.Data.(string))
		assert.True(t, sub2.WouldWait(ctx), "should still advertise it would wait after done")
	case <-time.After(1 * time.Second):
		assert.Fail(t, "took too long to complete")
	}

	cancel()
	assert.False(t, sub2.WouldWait(ctx), "should not wait after ctx is done")
}

func Test_promiseGetter_other_instance(t *testing.T) {
	client := cache.NewMemoryClient()
	ctx := context.Background()
	store := New(client, time.Second)
	store.Tomb().Go(store.Run)

	sub, prov := store.GetOrLock(ctx, "foo")

	assert.Nil(t, sub, "should not have a getter")
	assert.IsType(t, &setter{}, prov, "should have a setter")

	// New store to mimic a new instance
	store2 := New(client, time.Second)
	store2.Tomb().Go(store.Run)

	sub2, prov2 := store2.GetOrLock(ctx, "foo")
	assert.IsType(t, &pollGetter{}, sub2, "should have a getter")
	assert.IsType(t, &setter{}, prov2, "should have a setter")

	sub3, prov3 := store2.GetOrLock(ctx, "foo")
	assert.IsType(t, &promiseGetter{}, sub3, "should have a getter")
	assert.IsType(t, &setter{}, prov3, "should have a setter")

	go func() {
		_, err := sub2.Wait(ctx)
		assert.NoError(t, err)
	}()

	done := make(chan *cache.Item)
	go func() {
		item, err := sub3.Wait(ctx)
		assert.NoError(t, err)
		done <- item
	}()
	<-time.After(100 * time.Millisecond)

	err := prov.Done(ctx, "bar", 10*time.Second)
	assert.NoError(t, err)

	select {
	case item := <-done:
		assert.IsType(t, &cache.Item{}, item)
		assert.Equal(t, "bar", item.Data.(string))
	case <-time.After(1 * time.Second):
		assert.Fail(t, "took too long to complete")
	}
}
