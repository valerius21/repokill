// Package worker provides a generic worker pool implementation for concurrent operations.
package worker

import (
	"context"
	"sync"
)

const (
	// DefaultConcurrency is the default number of workers in the pool.
	DefaultConcurrency = 5
	// MaxConcurrency is the maximum allowed number of workers.
	MaxConcurrency = 20
	// DefaultQueueSize is the default buffer size for the job queue.
	DefaultQueueSize = 100
)

// Job represents a unit of work that can be executed by the worker pool.
type Job interface {
	// Execute performs the job's work and returns any error that occurred.
	Execute(ctx context.Context) error
	// Name returns a human-readable identifier for the job.
	Name() string
}

// Result represents the outcome of a job execution.
type Result struct {
	// JobName is the name of the job that was executed.
	JobName string
	// Error contains any error that occurred during execution, or nil on success.
	Error error
}

// WorkerPool manages a pool of workers that process jobs concurrently.
type WorkerPool struct {
	workers    int
	jobs       chan Job
	results    chan Result
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	started    bool
	stopped    bool
	jobsClosed bool
	mu         sync.Mutex
}

// Option is a function that configures the worker pool.
type Option func(*WorkerPool)

// WithConcurrency sets the number of workers in the pool.
// If n <= 0, DefaultConcurrency is used.
// If n > MaxConcurrency, MaxConcurrency is used.
func WithConcurrency(n int) Option {
	return func(p *WorkerPool) {
		if n <= 0 {
			p.workers = DefaultConcurrency
		} else if n > MaxConcurrency {
			p.workers = MaxConcurrency
		} else {
			p.workers = n
		}
	}
}

// WithQueueSize sets the buffer size for the job queue.
// If size <= 0, DefaultQueueSize is used.
func WithQueueSize(size int) Option {
	return func(p *WorkerPool) {
		if size <= 0 {
			size = DefaultQueueSize
		}
		p.jobs = make(chan Job, size)
	}
}

// NewWorkerPool creates a new worker pool with the given options.
// The pool is created but not started; call Start to begin processing jobs.
func NewWorkerPool(ctx context.Context, opts ...Option) *WorkerPool {
	childCtx, cancel := context.WithCancel(ctx)

	p := &WorkerPool{
		workers: DefaultConcurrency,
		jobs:    make(chan Job, DefaultQueueSize),
		results: make(chan Result, DefaultQueueSize),
		ctx:     childCtx,
		cancel:  cancel,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Start begins the worker goroutines. This method is idempotent;
// calling it multiple times has no effect after the first call.
func (p *WorkerPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return
	}
	p.started = true

	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

// worker is the main goroutine that processes jobs from the queue.
func (p *WorkerPool) worker() {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			return
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			result := Result{
				JobName: job.Name(),
			}

			// Execute the job, checking for context cancellation
			err := job.Execute(p.ctx)
			result.Error = err

			// Send result, but don't block if context is done
			select {
			case <-p.ctx.Done():
				return
			case p.results <- result:
			}
		}
	}
}

// Submit adds a job to the pool for processing.
// If the pool hasn't been started, it will be started automatically.
// Returns an error if the context is cancelled.
func (p *WorkerPool) Submit(job Job) error {
	p.mu.Lock()
	if !p.started {
		p.started = true
		for i := 0; i < p.workers; i++ {
			p.wg.Add(1)
			go p.worker()
		}
	}
	p.mu.Unlock()

	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case p.jobs <- job:
		return nil
	}
}

// Results returns the channel from which job results can be read.
func (p *WorkerPool) Results() <-chan Result {
	return p.results
}

// Stop gracefully shuts down the worker pool.
// It stops accepting new jobs and waits for all workers to finish.
// After Stop is called, the results channel is closed.
// Stop is idempotent; calling it multiple times has no effect.
func (p *WorkerPool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.started || p.stopped {
		return
	}
	p.stopped = true

	// Close the jobs channel to signal workers to stop (if not already closed)
	if !p.jobsClosed {
		close(p.jobs)
		p.jobsClosed = true
	}

	// Wait for all workers to finish
	p.wg.Wait()

	// Cancel the context
	p.cancel()

	// Close the results channel
	close(p.results)
}

// Wait blocks until all submitted jobs have been processed and the pool is stopped.
// This is useful for waiting on a batch of jobs to complete.
func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

// ProcessBatch submits all jobs and returns all results once complete.
// This is a convenience method for processing a batch of jobs synchronously.
// The pool is started and stopped automatically.
func (p *WorkerPool) ProcessBatch(ctx context.Context, jobs []Job) []Result {
	if len(jobs) == 0 {
		return nil
	}

	p.Start()

	// Submit all jobs
	go func() {
		for _, job := range jobs {
			if err := p.Submit(job); err != nil {
				// Context cancelled, stop submitting
				break
			}
		}
		// Mark jobs channel as closed
		p.mu.Lock()
		if !p.jobsClosed {
			close(p.jobs)
			p.jobsClosed = true
		}
		p.mu.Unlock()
	}()

	// Collect results
	results := make([]Result, 0, len(jobs))
	for i := 0; i < len(jobs); i++ {
		select {
		case <-ctx.Done():
			p.Stop()
			return results
		case result, ok := <-p.results:
			if !ok {
				return results
			}
			results = append(results, result)
		}
	}

	// Pool finished all jobs - clean up
	p.Stop()

	return results
}

// WorkerCount returns the number of workers in the pool.
func (p *WorkerPool) WorkerCount() int {
	return p.workers
}
