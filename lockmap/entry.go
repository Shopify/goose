package lockmap

import "sync"

type entry struct {
	// Used to make sure a channel is only closed once.
	once    sync.Once
	promise chan struct{}

	// Use int64 instead of time.Time for reduced memory usage
	expiration int64
}

func (e *entry) resolve() {
	e.once.Do(func() {
		close(e.promise)
	})
}
