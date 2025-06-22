package async

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewWorkerPool(t *testing.T) {
	t.Run("creates worker pool with specified size", func(t *testing.T) {
		pool := NewWorkerPool(5, 10)
		defer pool.Close()

		assert.Equal(t, 5, pool.size)
	})

	t.Run("uses minimum 1 worker", func(t *testing.T) {
		pool := NewWorkerPool(0, 10)
		defer pool.Close()

		assert.Equal(t, 1, pool.size)
	})
}

func TestWorkerPool_Submit(t *testing.T) {
	t.Run("processes tasks concurrently", func(t *testing.T) {
		pool := NewWorkerPool(3, 10)
		defer pool.Close()

		var mu sync.Mutex
		var counter int
		var wg sync.WaitGroup

		// Submit 10 tasks that increment a counter
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := pool.Submit(func() (interface{}, error) {
					mu.Lock()
					counter++
					mu.Unlock()
					return nil, nil
				})
				assert.NoError(t, err)
			}()
		}

		wg.Wait()

		// Wait for all tasks to complete
		time.Sleep(100 * time.Millisecond)
		assert.Equal(t, 10, counter)
	})

	t.Run("returns error when pool is closed", func(t *testing.T) {
		pool := NewWorkerPool(1, 1)
		pool.Close()

		err := pool.Submit(func() (interface{}, error) { return nil, nil })
		assert.Equal(t, ErrWorkerPoolClosed, err)
	})
}

func TestWorkerPool_SubmitWithContext(t *testing.T) {
	t.Run("respects context cancellation", func(t *testing.T) {
		pool := NewWorkerPool(1, 1)
		defer pool.Close()

		// Fill up the worker and buffer
		pool.Submit(func() (interface{}, error) {
			time.Sleep(100 * time.Millisecond)
			return nil, nil
		})


		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := pool.SubmitWithContext(ctx, func() (interface{}, error) {
			return nil, nil
		})

		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestWorkerPool_Results(t *testing.T) {
	t.Run("collects results from tasks", func(t *testing.T) {
		pool := NewWorkerPool(2, 5)
		defer pool.Close()

		// Submit tasks
		for i := 0; i < 3; i++ {
			i := i
			pool.Submit(func() (interface{}, error) {
				return i * 2, nil
			})
		}

		// Submit a task that returns an error
		errTest := errors.New("test error")
		pool.Submit(func() (interface{}, error) {
			return nil, errTest
		})

		// Collect results
		results := make([]Result, 0, 4)
		for i := 0; i < 4; i++ {
			select {
			case result := <-pool.Results():
				results = append(results, result)
			case <-time.After(100 * time.Millisecond):
				t.Fatal("timeout waiting for results")
			}
		}

		// Verify results
		assert.Len(t, results, 4)
		values := make([]int, 0, 3)
		var foundError bool
		for _, r := range results {
			if r.Err != nil {
				assert.Equal(t, errTest, r.Err)
				foundError = true
			} else {
				values = append(values, r.Value.(int))
			}
		}
		assert.True(t, foundError)
		assert.ElementsMatch(t, []int{0, 2, 4}, values)
	})
}

func TestWorkerPool_Close(t *testing.T) {
	t.Run("waits for pending tasks to complete", func(t *testing.T) {
		pool := NewWorkerPool(1, 1)

		// Submit a long-running task
		taskStarted := make(chan struct{})
		taskDone := make(chan struct{})

		pool.Submit(func() (interface{}, error) {
			close(taskStarted)
			time.Sleep(100 * time.Millisecond)
			close(taskDone)
			return nil, nil
		})


		// Wait for task to start
		<-taskStarted

		// Close should wait for the task to complete
		closeDone := make(chan struct{})
		go func() {
			pool.Close()
			close(closeDone)
		}()

		// Verify Close is blocked until task is done
		select {
		case <-closeDone:
			t.Fatal("Close returned before task was done")
		case <-time.After(50 * time.Millisecond):
			// Expected: Close is still waiting
		}

		// Wait for task to complete
		<-taskDone

		// Now Close should complete
		select {
		case <-closeDone:
			// Success
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Close didn't complete after task was done")
		}
	})

}

func TestWorkerPool_ConcurrentUsage(t *testing.T) {
	pool := NewWorkerPool(10, 100)
	defer pool.Close()

	const numTasks = 100
	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Start result consumer
	go func() {
		for result := range pool.Results() {
			if result.Err == nil {
				mu.Lock()
				counter++
				mu.Unlock()
			}
			wg.Done()
		}
	}()

	// Submit tasks
	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(n int) {
			err := pool.Submit(func() (interface{}, error) {
				time.Sleep(time.Duration(n%10) * time.Millisecond)
				return n * 2, nil
			})
			if err != nil {
				wg.Done()
			}
		}(i)
	}

	// Wait for all tasks to complete
	wg.Wait()

	// Give some time for the counter to update
	time.Sleep(50 * time.Millisecond)


	assert.Equal(t, numTasks, counter)
}

func TestWorkerPool_SubmitAfterClose(t *testing.T) {
	pool := NewWorkerPool(1, 1)
	pool.Close()

	err := pool.Submit(func() (interface{}, error) { return nil, nil })
	assert.Equal(t, ErrWorkerPoolClosed, err)

	err = pool.SubmitWithContext(context.Background(), func() (interface{}, error) {
		return nil, nil
	})
	assert.Equal(t, ErrWorkerPoolClosed, err)
}
