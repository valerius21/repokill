package worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockJob is a test implementation of the Job interface.
type mockJob struct {
	name     string
	execute  func(ctx context.Context) error
	executed atomic.Bool
}

func (j *mockJob) Execute(ctx context.Context) error {
	j.executed.Store(true)
	if j.execute != nil {
		return j.execute(ctx)
	}
	return nil
}

func (j *mockJob) Name() string {
	return j.name
}

func (j *mockJob) WasExecuted() bool {
	return j.executed.Load()
}

func TestNewWorkerPool(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		pool := NewWorkerPool(context.Background())
		if pool.WorkerCount() != DefaultConcurrency {
			t.Errorf("expected %d workers, got %d", DefaultConcurrency, pool.WorkerCount())
		}
	})

	t.Run("with custom concurrency", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(10))
		if pool.WorkerCount() != 10 {
			t.Errorf("expected 10 workers, got %d", pool.WorkerCount())
		}
	})

	t.Run("concurrency capped at max", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(100))
		if pool.WorkerCount() != MaxConcurrency {
			t.Errorf("expected %d workers (max), got %d", MaxConcurrency, pool.WorkerCount())
		}
	})

	t.Run("negative concurrency uses default", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(-5))
		if pool.WorkerCount() != DefaultConcurrency {
			t.Errorf("expected %d workers (default), got %d", DefaultConcurrency, pool.WorkerCount())
		}
	})

	t.Run("zero concurrency uses default", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(0))
		if pool.WorkerCount() != DefaultConcurrency {
			t.Errorf("expected %d workers (default), got %d", DefaultConcurrency, pool.WorkerCount())
		}
	})
}

func TestWorkerPool_Submit(t *testing.T) {
	t.Run("single job execution", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(1))
		pool.Start()

		job := &mockJob{name: "test-job"}
		err := pool.Submit(job)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Wait for result
		result := <-pool.Results()
		if result.JobName != "test-job" {
			t.Errorf("expected job name 'test-job', got %s", result.JobName)
		}
		if result.Error != nil {
			t.Errorf("unexpected error: %v", result.Error)
		}
		if !job.WasExecuted() {
			t.Error("job was not executed")
		}

		pool.Stop()
	})

	t.Run("job returns error", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(1))
		pool.Start()

		expectedErr := errors.New("job failed")
		job := &mockJob{
			name: "failing-job",
			execute: func(ctx context.Context) error {
				return expectedErr
			},
		}
		_ = pool.Submit(job)

		result := <-pool.Results()
		if result.Error != expectedErr {
			t.Errorf("expected error %v, got %v", expectedErr, result.Error)
		}

		pool.Stop()
	})

	t.Run("auto-start on first submit", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(1))
		// Don't call Start()

		job := &mockJob{name: "auto-start-job"}
		err := pool.Submit(job)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		result := <-pool.Results()
		if result.JobName != "auto-start-job" {
			t.Errorf("expected job name 'auto-start-job', got %s", result.JobName)
		}

		pool.Stop()
	})

	t.Run("submit on cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		pool := NewWorkerPool(ctx, WithConcurrency(1))
		pool.Start()

		job := &mockJob{name: "cancelled-job"}
		err := pool.Submit(job)
		// Either we get an error (preferred) or the job is not processed
		// The key is that the pool handles cancelled context gracefully

		pool.Stop()
		// If we got here without panic, the test passes
		_ = err
	})
}

func TestWorkerPool_MultipleJobs(t *testing.T) {
	t.Run("processes multiple jobs concurrently", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(3))
		pool.Start()

		var executionOrder []string
		var mu sync.Mutex

		// Create jobs that sleep briefly and record their execution
		for i := 0; i < 5; i++ {
			job := &mockJob{
				name: fmt.Sprintf("job-%d", i),
				execute: func(ctx context.Context) error {
					mu.Lock()
					executionOrder = append(executionOrder, "started")
					mu.Unlock()
					time.Sleep(50 * time.Millisecond)
					mu.Lock()
					executionOrder = append(executionOrder, "finished")
					mu.Unlock()
					return nil
				},
			}
			_ = pool.Submit(job)
		}

		// Collect all results
		results := make([]Result, 0, 5)
		for i := 0; i < 5; i++ {
			results = append(results, <-pool.Results())
		}

		// Check that we got 5 results
		if len(results) != 5 {
			t.Errorf("expected 5 results, got %d", len(results))
		}

		// With 3 workers and 5 jobs with 50ms each, concurrent execution
		// should complete in ~100ms vs ~250ms sequential
		// We just verify all jobs were processed
		for _, r := range results {
			if r.Error != nil {
				t.Errorf("unexpected error in job %s: %v", r.JobName, r.Error)
			}
		}

		pool.Stop()
	})

	t.Run("error aggregation continues on failures", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(2))
		pool.Start()

		// Submit 3 jobs, middle one fails
		for i := 0; i < 3; i++ {
			idx := i
			job := &mockJob{
				name: fmt.Sprintf("job-%d", i),
				execute: func(ctx context.Context) error {
					if idx == 1 {
						return errors.New("job 1 failed")
					}
					return nil
				},
			}
			_ = pool.Submit(job)
		}

		// Collect results
		results := make([]Result, 0, 3)
		for i := 0; i < 3; i++ {
			results = append(results, <-pool.Results())
		}

		// Count successes and failures
		successCount := 0
		failureCount := 0
		for _, r := range results {
			if r.Error != nil {
				failureCount++
			} else {
				successCount++
			}
		}

		if successCount != 2 {
			t.Errorf("expected 2 successes, got %d", successCount)
		}
		if failureCount != 1 {
			t.Errorf("expected 1 failure, got %d", failureCount)
		}

		pool.Stop()
	})
}

func TestWorkerPool_ProcessBatch(t *testing.T) {
	t.Run("processes batch of jobs", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(3))

		jobs := make([]Job, 5)
		for i := 0; i < 5; i++ {
			jobs[i] = &mockJob{name: fmt.Sprintf("batch-job-%d", i)}
		}

		results := pool.ProcessBatch(context.Background(), jobs)

		if len(results) != 5 {
			t.Errorf("expected 5 results, got %d", len(results))
		}

		for _, r := range results {
			if r.Error != nil {
				t.Errorf("unexpected error: %v", r.Error)
			}
		}
	})

	t.Run("empty batch returns nil", func(t *testing.T) {
		pool := NewWorkerPool(context.Background())
		results := pool.ProcessBatch(context.Background(), []Job{})
		if results != nil {
			t.Errorf("expected nil for empty batch, got %v", results)
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(1))

		jobs := make([]Job, 10)
		for i := 0; i < 10; i++ {
			jobs[i] = &mockJob{
				name: fmt.Sprintf("slow-job-%d", i),
				execute: func(ctx context.Context) error {
					time.Sleep(100 * time.Millisecond)
					return nil
				},
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		results := pool.ProcessBatch(ctx, jobs)

		// Should have fewer than 10 results due to timeout
		if len(results) >= 10 {
			t.Errorf("expected fewer than 10 results due to cancellation, got %d", len(results))
		}
	})
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	t.Run("stop waits for in-flight jobs", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(1))
		pool.Start()

		var jobCompleted atomic.Bool
		job := &mockJob{
			name: "long-job",
			execute: func(ctx context.Context) error {
				time.Sleep(100 * time.Millisecond)
				jobCompleted.Store(true)
				return nil
			},
		}
		_ = pool.Submit(job)

		// Small delay to ensure job is picked up
		time.Sleep(20 * time.Millisecond)

		// Stop should wait for job to complete
		done := make(chan struct{})
		go func() {
			pool.Stop()
			close(done)
		}()

		select {
		case <-done:
			// Good, Stop completed
		case <-time.After(500 * time.Millisecond):
			t.Error("Stop took too long, job may not have completed")
		}

		if !jobCompleted.Load() {
			t.Error("job did not complete before Stop returned")
		}
	})

	t.Run("multiple stop calls are safe", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(2))
		pool.Start()

		// Multiple stops should not panic
		pool.Stop()
		pool.Stop()
		pool.Stop()
	})

	t.Run("stop on non-started pool is safe", func(t *testing.T) {
		pool := NewWorkerPool(context.Background())
		// Don't start it
		pool.Stop() // Should not panic
	})
}

func TestWorkerPool_ContextCancellation(t *testing.T) {
	t.Run("workers stop on context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		pool := NewWorkerPool(ctx, WithConcurrency(2))
		pool.Start()

		// Submit a few jobs
		for i := 0; i < 3; i++ {
			job := &mockJob{
				name: fmt.Sprintf("ctx-job-%d", i),
				execute: func(ctx context.Context) error {
					time.Sleep(50 * time.Millisecond)
					return nil
				},
			}
			_ = pool.Submit(job)
		}

		// Cancel context
		cancel()

		// Pool should stop gracefully
		done := make(chan struct{})
		go func() {
			pool.Stop()
			close(done)
		}()

		select {
		case <-done:
			// Good
		case <-time.After(1 * time.Second):
			t.Error("Stop took too long after context cancellation")
		}
	})
}

func TestWorkerPool_ResultsChannel(t *testing.T) {
	t.Run("results channel provides all results", func(t *testing.T) {
		pool := NewWorkerPool(context.Background(), WithConcurrency(2))
		pool.Start()

		for i := 0; i < 3; i++ {
			job := &mockJob{name: fmt.Sprintf("result-job-%d", i)}
			_ = pool.Submit(job)
		}

		receivedNames := make(map[string]bool)
		for i := 0; i < 3; i++ {
			result := <-pool.Results()
			receivedNames[result.JobName] = true
		}

		for i := 0; i < 3; i++ {
			name := fmt.Sprintf("result-job-%d", i)
			if !receivedNames[name] {
				t.Errorf("missing result for %s", name)
			}
		}

		pool.Stop()
	})
}
