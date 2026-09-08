package registry

import "sync"

var (
	mu           sync.Mutex
	deregisterFn func() error
)

func SetDeregister(fn func() error) {
	mu.Lock()
	defer mu.Unlock()
	deregisterFn = fn
}

func Deregister() error {
	mu.Lock()
	fn := deregisterFn
	mu.Unlock()
	if fn == nil {
		return nil
	}
	return fn()
}
