package sharedstore

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/pkg/errors"

	"github.com/Shopify/goose/logger"
)

var log = logger.New("sharedstore")

type Item struct {
	Expiration time.Time
	Data       interface{}
}

func (i *Item) Bytes() ([]byte, error) {
	encoded := &bytes.Buffer{}
	enc := gob.NewEncoder(encoded)
	err := enc.Encode(*i)
	return encoded.Bytes(), errors.Wrap(err, "unable to encode item")
}

func itemFromBytes(b []byte) (*Item, error) {
	dec := gob.NewDecoder(bytes.NewReader(b))
	var item Item
	err := dec.Decode(&item)
	return &item, errors.Wrap(err, "unable to decode item")
}
