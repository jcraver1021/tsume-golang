package concurrency

import (
	"sync"
)

type PubSub[I any] struct {
	subscribers []chan I
	mu          sync.Mutex
}

func NewPubSub[I any]() *PubSub[I] {
	return &PubSub[I]{}
}

func (ps *PubSub[I]) Subscribe() <-chan I {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan I)
	ps.subscribers = append(ps.subscribers, ch)
	return ch
}

func (ps *PubSub[I]) Publish(msg I) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	var wg sync.WaitGroup

	for _, ch := range ps.subscribers {
		wg.Add(1)
		go func(ch chan I) {
			defer wg.Done()
			ch <- msg
		}(ch)
	}
	wg.Wait()
}

func (ps *PubSub[I]) Unsubscribe(ch <-chan I) {
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

func (ps *PubSub[I]) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, ch := range ps.subscribers {
		close(ch)
	}
	ps.subscribers = nil
}
