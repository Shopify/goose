package sharedstore

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Shopify/goose/logger"
)

func Test_condGetter(t *testing.T) {
	client := memoryClient{}
	ctx := logger.WithLoggable(context.Background(), &client)
	store := New(&client, time.Second)
	store.Tomb().Go(store.Run)

	sub, prov := store.GetOrLock(ctx, "foo")

	assert.Nil(t, sub, "should not have a getter")
	assert.IsType(t, &setter{}, prov, "should have a setter")

	sub2, prov2 := store.GetOrLock(ctx, "foo")
	assert.IsType(t, &promiseGetter{}, sub2, "should have a getter")
	assert.IsType(t, &setter{}, prov2, "should have a setter")

	done := make(chan *Item)
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
	case <-time.After(1 * time.Second):
		assert.Fail(t, "took too long to complete")
	}
}

func Test_lockGetter(t *testing.T) {
	client := memoryClient{}
	ctx := logger.WithLoggable(context.Background(), &client)
	store := New(&client, time.Second)
	store.Tomb().Go(store.Run)

	sub, prov := store.GetOrLock(ctx, "foo")

	assert.Nil(t, sub, "should not have a getter")
	assert.IsType(t, &setter{}, prov, "should have a setter")

	// New store to mimic a new instance
	store2 := New(&client, time.Second)
	store2.Tomb().Go(store.Run)

	sub2, prov2 := store2.GetOrLock(ctx, "foo")
	assert.IsType(t, &pollGetter{}, sub2, "should have a getter")
	assert.IsType(t, &setter{}, prov2, "should have a setter")

	done := make(chan *Item)
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
	case <-time.After(1 * time.Second):
		assert.Fail(t, "took too long to complete")
	}
}

func Test_condGetter_other_instance(t *testing.T) {
	client := memoryClient{}
	ctx := logger.WithLoggable(context.Background(), &client)
	store := New(&client, time.Second)
	store.Tomb().Go(store.Run)

	sub, prov := store.GetOrLock(ctx, "foo")

	assert.Nil(t, sub, "should not have a getter")
	assert.IsType(t, &setter{}, prov, "should have a setter")

	// New store to mimic a new instance
	store2 := New(&client, time.Second)
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

	done := make(chan *Item)
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
		assert.IsType(t, &Item{}, item)
		assert.Equal(t, "bar", item.Data.(string))
	case <-time.After(1 * time.Second):
		assert.Fail(t, "took too long to complete")
	}
}
