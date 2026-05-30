package concurrency_test

import (
	"testing"
	"time"

	. "tsumegolang/pkg/concurrency"
)

func TestPubSub(t *testing.T) {
	testCases := []struct {
		name string
		n    int
	}{
		{
			name: "1 subscriber",
			n:    1,
		},
		{
			name: "5 subscribers",
			n:    5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ps := NewPubSub[string]()
			subs := make([]<-chan string, tc.n)

			for i := 0; i < tc.n; i++ {
				subs[i] = ps.Subscribe()
			}

			go ps.Publish("hello")

			for i := 0; i < tc.n; i++ {
				select {
				case msg := <-subs[i]:
					if msg != "hello" {
						t.Errorf("expected 'hello', got '%s'", msg)
					}
				case <-time.After(1 * time.Second):
					t.Errorf("subscriber %d did not receive message", i)
				}
			}

			for i := 0; i < tc.n; i++ {
				ps.Unsubscribe(subs[i])
			}

			go ps.Publish("world")

			for i := 0; i < tc.n; i++ {
				select {
				case msg, ok := <-subs[i]:
					if ok {
						t.Errorf("subscriber %d should have been unsubscribed but received '%s'", i, msg)
					}
				case <-time.After(1 * time.Second):
					// Expected timeout since subscriber should be unsubscribed
				}
			}
		})
	}
}

func TestPubSubClose(t *testing.T) {
	ps := NewPubSub[string]()
	sub := ps.Subscribe()

	go ps.Publish("hello")
	select {
	case msg := <-sub:
		if msg != "hello" {
			t.Errorf("expected 'hello', got '%s'", msg)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("subscriber did not receive message")
	}

	ps.Close()

	go ps.Publish("world")
	select {
	case msg, ok := <-sub:
		if ok {
			t.Errorf("subscriber should have been closed but received '%s'", msg)
		}
	case <-time.After(1 * time.Second):
		// Expected timeout since subscriber should be closed
	}
}
