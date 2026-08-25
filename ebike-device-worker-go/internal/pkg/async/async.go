package async

import (
	"sync"
	"time"
)

var wg sync.WaitGroup

// Go runs fn in a goroutine tracked for graceful shutdown.
func Go(fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}

// Wait blocks until tracked goroutines finish or timeout elapses.
func Wait(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}
