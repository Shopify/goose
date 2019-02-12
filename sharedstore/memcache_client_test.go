package sharedstore

import (
	"encoding/gob"
	"testing"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/stretchr/testify/assert"
)

func ExampleNewMemcacheClient() {
	memcacheClient := memcache.New("localhost:11211")
	client := NewMemcacheClient(memcacheClient)

	_ = New(client, 10*time.Second)
}

func Test_encodeMemcacheItem(t *testing.T) {
	type test struct {
		Foo string
	}
	gob.Register(test{})

	tests := map[string]Item{
		"empty":      {},
		"expiration": {Expiration: time.Unix(10000, 0)},
		"struct":     {Data: test{Foo: "bar"}},
		"integer":    {Data: 123},
		"float":      {Data: 1.2},
		"string":     {Data: "123"},
		"nil":        {Data: nil},
	}

	for name, item := range tests {
		t.Run(name, func(t *testing.T) {
			enc, err := encodeMemcacheItem(name, &item)
			assert.NoError(t, err)
			assert.NotNil(t, enc)

			dec, err := decodeMemcacheItem(enc)
			assert.NoError(t, err)
			assert.NotNil(t, dec)

			assert.EqualValues(t, item, *dec)
		})
	}
}
