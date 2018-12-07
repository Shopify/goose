package sharedstore

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type Client interface {
	Get(key string) (*Item, error)
	Set(key string, item *Item) error
	Add(key string, item *Item) error
	Delete(key string) error
}

type memoryClient struct {
	data sync.Map
}

// NewMemoryClient returns a Client that only stores in memory.
// This defeats the purpose of the Store, but it’s useful for stubbing tests.
// Note that it does not honour expiration.
func NewMemoryClient() Client {
	return &memoryClient{}
}

func (c *memoryClient) Get(key string) (*Item, error) {
	item, ok := c.data.Load(key)
	if !ok {
		return nil, nil
	}
	return item.(*Item), nil

}

func (c *memoryClient) Set(key string, item *Item) error {
	c.data.Store(key, item)
	return nil
}

func (c *memoryClient) Add(key string, item *Item) error {
	_, loaded := c.data.LoadOrStore(key, item)
	if loaded {
		return ErrNotStored
	}
	return nil
}

func (c *memoryClient) Delete(key string) error {
	c.data.Delete(key)
	return nil
}

func (c *memoryClient) LogFields() logrus.Fields {
	fields := logrus.Fields{}
	c.data.Range(func(key, value interface{}) bool {
		m := value.(*Item)
		fields[key.(string)] = m.Data
		return true
	})
	return fields
}
