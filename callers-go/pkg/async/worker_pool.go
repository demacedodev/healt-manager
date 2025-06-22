package async

import (
	"context"
	"sync"
)

// Task represents a unit of work that returns a result and an error
type Task func() (interface{}, error)

// Result represents the output of a completed task
type Result struct {
	Value interface{}
	Err   error
}

// WorkerPool manages a pool of worker goroutines to process tasks
type WorkerPool struct {
	tasks    chan Task
	results  chan Result
	size     int
	wg       sync.WaitGroup
	stop     chan struct{}
	isClosed bool
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(workerCount, taskBufferSize int) *WorkerPool {
	if workerCount < 1 {
		workerCount = 1
	}

	wp := &WorkerPool{
		tasks:    make(chan Task, taskBufferSize),
		results:  make(chan Result, taskBufferSize),
		size:     workerCount,
		stop:     make(chan struct{}),
		isClosed: false,
	}

	wp.start()
	return wp
}

// Submit adds a task to the worker pool
func (wp *WorkerPool) Submit(task Task) error {
	if wp.isClosed {
		return ErrWorkerPoolClosed
	}

	wp.tasks <- task
	return nil
}

// SubmitWithContext adds a task to the worker pool with context support
func (wp *WorkerPool) SubmitWithContext(ctx context.Context, task Task) error {
	if wp.isClosed {
		return ErrWorkerPoolClosed
	}

	select {
	case wp.tasks <- task:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Results returns a channel to receive task results
func (wp *WorkerPool) Results() <-chan Result {
	return wp.results
}

// Close stops the worker pool and waits for all tasks to complete
func (wp *WorkerPool) Close() {
	if wp.isClosed {
		return
	}

	close(wp.stop)
	wp.wg.Wait()

	wp.isClosed = true
	close(wp.tasks)
	close(wp.results)
}

// start initializes the worker pool with the specified number of workers
func (wp *WorkerPool) start() {
	for i := 0; i < wp.size; i++ {
		wp.wg.Add(1)
		go wp.worker()
	}
}

// worker processes tasks from the queue
func (wp *WorkerPool) worker() {
	defer wp.wg.Done()

	for {
		select {
		case task, ok := <-wp.tasks:
			if !ok {
				return
			}

			result, err := task()
			wp.results <- Result{Value: result, Err: err}

		case <-wp.stop:
			return
		}
	}
}

// Errors
var (
	ErrWorkerPoolClosed = newError("worker pool is closed")
)

// newError creates a new error with the given message
func newError(message string) error {
	return &workerPoolError{message}
}

type workerPoolError struct {
	message string
}

func (e *workerPoolError) Error() string {
	return e.message
}
