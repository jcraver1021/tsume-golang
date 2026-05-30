package concurrency

import (
	"sync"
)

type WorkFn[S, T any] func(T) (S, error)

type WorkError[T any] struct {
	Input T
	Err   error
}

func FanOut[S, T any](in <-chan T, n int, fn WorkFn[S, T]) ([]<-chan S, <-chan WorkError[T]) {
	outs := make([]<-chan S, n)
	errs := make(chan WorkError[T])
	var wg sync.WaitGroup

	for i := range n {
		out := make(chan S)
		outs[i] = out
		wg.Add(1)

		go func(out chan S) {
			defer wg.Done()
			defer close(out)
			for v := range in {
				result, err := fn(v)
				if err != nil {
					errs <- WorkError[T]{Input: v, Err: err}
				} else {
					out <- result
				}
			}
		}(out)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	return outs, errs
}

func FanIn[T any](channels ...<-chan T) <-chan T {
	out := make(chan T)
	var wg sync.WaitGroup

	for _, in := range channels {
		wg.Add(1)
		go func(c <-chan T) {
			defer wg.Done()
			for v := range c {
				out <- v
			}
		}(in)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
