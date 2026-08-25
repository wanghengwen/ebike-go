package workerpool

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPoolRunsAllJobs(t *testing.T) {
	p := New(4, 8)
	var count int64
	var wg sync.WaitGroup
	const n = 100
	wg.Add(n)
	for i := 0; i < n; i++ {
		p.Submit(func() {
			atomic.AddInt64(&count, 1)
			wg.Done()
		})
	}
	wg.Wait()
	p.Close()
	if got := atomic.LoadInt64(&count); got != n {
		t.Fatalf("expected %d jobs run, got %d", n, got)
	}
}

func TestPoolBoundsConcurrency(t *testing.T) {
	const workers = 3
	p := New(workers, 0)
	defer p.Close()

	var inFlight, maxInFlight int64
	var wg sync.WaitGroup
	const n = 30
	wg.Add(n)
	for i := 0; i < n; i++ {
		p.Submit(func() {
			cur := atomic.AddInt64(&inFlight, 1)
			for {
				old := atomic.LoadInt64(&maxInFlight)
				if cur <= old || atomic.CompareAndSwapInt64(&maxInFlight, old, cur) {
					break
				}
			}
			time.Sleep(5 * time.Millisecond)
			atomic.AddInt64(&inFlight, -1)
			wg.Done()
		})
	}
	wg.Wait()
	if max := atomic.LoadInt64(&maxInFlight); max > workers {
		t.Fatalf("concurrency exceeded pool size: max=%d workers=%d", max, workers)
	}
}

func TestPoolCloseIsIdempotent(t *testing.T) {
	p := New(2, 2)
	p.Submit(func() {})
	p.Close()
	p.Close() // must not panic
}

func TestPoolRecoversJobPanic(t *testing.T) {
	p := New(2, 4)
	defer p.Close()

	// A panicking job must not crash the process nor kill the worker.
	var done sync.WaitGroup
	done.Add(1)
	p.Submit(func() { panic("boom") })

	var ran int64
	p.Submit(func() {
		atomic.AddInt64(&ran, 1)
		done.Done()
	})
	done.Wait()
	if atomic.LoadInt64(&ran) != 1 {
		t.Fatalf("pool did not keep processing after a panicking job")
	}
}

func TestPoolSubmitAfterCloseDoesNotPanic(t *testing.T) {
	p := New(2, 2)
	p.Close()
	// Submitting after close must drop silently, never send-on-closed panic.
	p.Submit(func() { t.Fatalf("job should not run after close") })
}

func TestPoolCloseWithBlockedSubmit(t *testing.T) {
	// queueSize 0 + occupied workers → Submit blocks; Close must unblock it
	// without a send-on-closed-channel panic.
	release := make(chan struct{})
	p := New(1, 0)
	p.Submit(func() { <-release }) // occupies the single worker

	submitted := make(chan struct{})
	go func() {
		p.Submit(func() {}) // blocks until Close signals done
		close(submitted)
	}()

	time.Sleep(20 * time.Millisecond) // let the goroutine reach a blocking Submit
	close(release)
	p.Close()
	select {
	case <-submitted:
	case <-time.After(time.Second):
		t.Fatalf("blocked Submit was not released by Close")
	}
}
