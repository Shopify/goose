package sharedstore

import (
	"time"

	"github.com/Shopify/goose/logger"
)

var log = logger.New("sharedstore")

type Item struct {
	Expiration time.Time
	Data       interface{}
}
