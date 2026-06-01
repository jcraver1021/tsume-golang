package concurrency

import (
	"sync"
)

func FanOut[I, O any](in <-chan I, n int, fn Job[I, O]) []<-chan JobResult[I, O] {
	outs := make([]<-chan JobResult[I, O], n)

	for i := range outs {
		out := make(chan JobResult[I, O])
		outs[i] = out

		go func(out chan JobResult[I, O]) {
			defer close(out)

			for v := range in {
				y := fn(v)
				out <- y
			}
		}(out)
	}

	return outs
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
