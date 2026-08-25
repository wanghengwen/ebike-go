package pool

import (
	"log"
	"sync"
)

// WorkerPool implements a fixed-size goroutine pool with a bounded task queue.
// When the queue is full, tasks are executed synchronously in the caller's
// goroutine (CallerRunsPolicy), providing natural back-pressure.
type WorkerPool struct {
	tasks chan func()
	wg    sync.WaitGroup
}

// NewWorkerPool creates a pool with maxWorkers goroutines and a buffered
// task channel of the given queueSize.
func NewWorkerPool(maxWorkers, queueSize int) *WorkerPool {
	p := &WorkerPool{
		tasks: make(chan func(), queueSize),
	}

	p.wg.Add(maxWorkers)
	for i := 0; i < maxWorkers; i++ {
		go func(workerID int) {
			defer p.wg.Done()
			for task := range p.tasks {
				func() {
					defer func() {
						if r := recover(); r != nil {
							log.Printf("[WorkerPool] worker %d recovered from panic: %v", workerID, r)
						}
					}()
					task()
				}()
			}
		}(i)
	}

	log.Printf("[WorkerPool] started with %d workers, queue size %d", maxWorkers, queueSize)
	return p
}

// Submit enqueues a task for asynchronous execution. If the queue is full,
// the task is executed synchronously in the caller's goroutine (CallerRunsPolicy).
func (p *WorkerPool) Submit(task func()) {
	select {
	case p.tasks <- task:
		// enqueued successfully
	default:
		// queue full – execute in caller goroutine (CallerRunsPolicy)
		log.Printf("[WorkerPool] queue full, executing task in caller goroutine")
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[WorkerPool] caller goroutine recovered from panic: %v", r)
				}
			}()
			task()
		}()
	}
}

// Shutdown closes the task channel and blocks until all workers finish
// their current tasks.
func (p *WorkerPool) Shutdown() {
	close(p.tasks)
	p.wg.Wait()
	log.Printf("[WorkerPool] shutdown complete")
}
