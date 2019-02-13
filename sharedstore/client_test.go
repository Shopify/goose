package sharedstore

import (
	"time"
)

func ExampleNewMemoryClient() {
	client := NewMemoryClient()

	_ = New(client, 10*time.Second)
}
