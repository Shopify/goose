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

	var data string
	sub, prov := store.GetOrLock(ctx, "foo", &data)

	assert.Nil(t, sub, "should not have a getter")
	assert.IsType(t, &setter{}, prov, "should have a setter")

	var data2 string
	sub2, prov2 := store.GetOrLock(ctx, "foo", &data2)
	assert.IsType(t, &promiseGetter{}, sub2, "should have a getter")
	assert.True(t, sub2.WouldWait(ctx))
	assert.IsType(t, &setter{}, prov2, "should have a setter")

	done := make(chan error)
	go func() {
		done <- sub2.Wait(ctx)
	}()
	<-time.After(100 * time.Millisecond)

	err := prov.Done(ctx, "bar", 10*time.Second)
	assert.NoError(t, err)

	select {
	case err := <-done:
		assert.NoError(t, err)
		assert.Equal(t, "bar", data2)
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

	var data string
	sub, prov := store.GetOrLock(ctx, "foo", &data)

	assert.Nil(t, sub, "should not have a getter")
	assert.IsType(t, &setter{}, prov, "should have a setter")

	// New store to mimic a new instance
	store2 := New(client, time.Second)
	store2.Tomb().Go(store.Run)

	var data2 string
	sub2, prov2 := store2.GetOrLock(ctx, "foo", &data2)
	assert.IsType(t, &pollGetter{}, sub2, "should have a getter")
	assert.True(t, sub2.WouldWait(ctx))
	assert.IsType(t, &setter{}, prov2, "should have a setter")

	done := make(chan error)
	go func() {
		done <- sub2.Wait(ctx)
	}()
	<-time.After(100 * time.Millisecond)

	err := prov.Done(ctx, "bar", 10*time.Second)
	assert.NoError(t, err)

	select {
	case err := <-done:
		assert.NoError(t, err)
		assert.Equal(t, "bar", data2)
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

	var data string
	sub, prov := store.GetOrLock(ctx, "foo", &data)

	assert.Nil(t, sub, "should not have a getter")
	assert.IsType(t, &setter{}, prov, "should have a setter")

	// New store to mimic a new instance
	store2 := New(client, time.Second)
	store2.Tomb().Go(store.Run)

	var data2 string
	sub2, prov2 := store2.GetOrLock(ctx, "foo", &data2)
	assert.IsType(t, &pollGetter{}, sub2, "should have a getter")
	assert.IsType(t, &setter{}, prov2, "should have a setter")

	var data3 string
	sub3, prov3 := store2.GetOrLock(ctx, "foo", &data3)
	assert.IsType(t, &promiseGetter{}, sub3, "should have a getter")
	assert.IsType(t, &setter{}, prov3, "should have a setter")

	go func() {
		err := sub2.Wait(ctx)
		assert.NoError(t, err)
	}()

	done := make(chan error)
	go func() {
		done <- sub3.Wait(ctx)
	}()
	<-time.After(100 * time.Millisecond)

	err := prov.Done(ctx, "bar", 10*time.Second)
	assert.NoError(t, err)

	select {
	case err := <-done:
		assert.NoError(t, err)
		assert.Equal(t, "bar", data3)
	case <-time.After(1 * time.Second):
		assert.Fail(t, "took too long to complete")
	}
}
