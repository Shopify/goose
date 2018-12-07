package sharedstore

import (
	"encoding/gob"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestData_Bytes(t *testing.T) {
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

	for name, c := range tests {
		t.Run(name, func(t *testing.T) {
			b, err := c.Bytes()
			assert.NoError(t, err)
			assert.NotNil(t, b)

			c2, err := itemFromBytes(b)
			assert.NoError(t, err)
			assert.NotNil(t, c2)

			assert.EqualValues(t, c, *c2)
		})
	}
}
