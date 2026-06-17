package concurrency

import (
	"context"
	"sync"
)

type workerRequest[I any, O any] struct {
	input    I
	resultCh chan JobResult[I, O]
}

type WorkerPool[I any, O any] struct {
	job        Job[I, O]
	numWorkers int
	queue      chan workerRequest[I, O]
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewWorkerPool[I any, O any](job Job[I, O], numWorkers int) *WorkerPool[I, O] {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool[I, O]{
		job:        job,
		numWorkers: numWorkers,
		queue:      make(chan workerRequest[I, O], numWorkers),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (wp *WorkerPool[I, O]) Start() {
	for range cap(wp.queue) {
		wp.wg.Add(1)
		go wp.worker()
	}
}

func (wp *WorkerPool[I, O]) worker() {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			// Drain remaining requests and close their result channels
			for req := range wp.queue {
				close(req.resultCh)
			}
			return
		case req, ok := <-wp.queue:
			if !ok {
				return
			}
			result := wp.job(req.input)
			req.resultCh <- result
			close(req.resultCh)
		}
	}
}

func (wp *WorkerPool[I, O]) Submit(input I) (chan JobResult[I, O], error) {
	resultCh := make(chan JobResult[I, O], 1)
	select {
	case wp.queue <- workerRequest[I, O]{input: input, resultCh: resultCh}:
		return resultCh, nil
	case <-wp.ctx.Done():
		return nil, wp.ctx.Err()
	}
}

func (wp *WorkerPool[I, O]) Shutdown() {
	close(wp.queue)
	wp.cancel()
	wp.wg.Wait()
}
