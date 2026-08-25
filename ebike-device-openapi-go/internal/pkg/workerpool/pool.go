// Package workerpool provides a simple bounded worker pool used to apply
// backpressure and bound concurrency when processing inbound MQTT messages.
package workerpool

import (
	"sync"

	"ebike-device-openapi-go/internal/pkg/logger"
	"go.uber.org/zap"
)

// Job is a unit of work executed by a pool worker.
type Job func()

// Pool runs jobs on a fixed number of workers backed by a bounded queue.
// Submit blocks when the queue is full, propagating backpressure to the caller.
//
// The job channel is intentionally never closed: shutdown is signalled via a
// separate done channel and Submit selects on it, so a concurrent Submit can never
// panic with "send on closed channel". Each job runs with panic recovery so a bad
// message can never crash the process.
type Pool struct {
	jobs      chan Job
	done      chan struct{}
	wg        sync.WaitGroup
	closeOnce sync.Once
}

// New creates and starts a pool with the given worker count and queue capacity.
// workers < 1 is clamped to 1; queueSize < 0 is clamped to 0 (fully synchronous).
func New(workers, queueSize int) *Pool {
	if workers < 1 {
		workers = 1
	}
	if queueSize < 0 {
		queueSize = 0
	}
	p := &Pool{
		jobs: make(chan Job, queueSize),
		done: make(chan struct{}),
	}
	p.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go p.worker()
	}
	return p
}

func (p *Pool) worker() {
	defer p.wg.Done()
	for {
		select {
		case job := <-p.jobs:
			runJob(job)
		case <-p.done:
			// Drain any queued jobs, then exit.
			for {
				select {
				case job := <-p.jobs:
					runJob(job)
				default:
					return
				}
			}
		}
	}
}

// runJob executes a single job, recovering from panics so a single malformed
// message (e.g. a decoder type assertion) cannot crash the whole process.
func runJob(job Job) {
	if job == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("workerpool job panic recovered", zap.Any("recover", r))
		}
	}()
	job()
}

// Submit enqueues a job, blocking while the bounded queue is full. This blocking is
// the backpressure signal: the caller (MQTT dispatch goroutine) cannot outrun the
// pool's processing capacity. Once the pool is closing, the job is dropped instead
// of blocking forever.
func (p *Pool) Submit(job Job) {
	if p == nil || job == nil {
		return
	}
	select {
	case p.jobs <- job:
	case <-p.done:
	}
}

// Close signals shutdown and waits for all in-flight and queued jobs to finish.
// Safe to call multiple times.
func (p *Pool) Close() {
	if p == nil {
		return
	}
	p.closeOnce.Do(func() { close(p.done) })
	p.wg.Wait()
}
