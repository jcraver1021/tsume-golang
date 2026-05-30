package concurrency

import (
	"sync"
)

type PubSub[T any] struct {
	subscribers []chan T
	mu          sync.Mutex
}

func NewPubSub[T any]() *PubSub[T] {
	return &PubSub[T]{}
}

func (ps *PubSub[T]) Subscribe() <-chan T {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan T)
	ps.subscribers = append(ps.subscribers, ch)
	return ch
}

func (ps *PubSub[T]) Publish(msg T) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	var wg sync.WaitGroup

	for _, ch := range ps.subscribers {
		wg.Add(1)
		go func(ch chan T) {
			defer wg.Done()
			ch <- msg
		}(ch)
	}
	wg.Wait()
}

func (ps *PubSub[T]) Unsubscribe(ch <-chan T) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for i, subscriber := range ps.subscribers {
		if subscriber == ch {
			ps.subscribers = append(ps.subscribers[:i], ps.subscribers[i+1:]...)
			close(subscriber)
			break
		}
	}
}

func (ps *PubSub[T]) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, ch := range ps.subscribers {
		close(ch)
	}
	ps.subscribers = nil
}
