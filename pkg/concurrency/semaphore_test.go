package concurrency_test

import (
	"testing"

	. "tsumegolang/pkg/concurrency"
)

func TestSemaphore(t *testing.T) {
	testCases := []struct {
		name     string
		capacity int
	}{
		{
			name:     "Semaphore with capacity 1",
			capacity: 1,
		},
		{
			name:     "Semaphore with capacity 2",
			capacity: 2,
		},
		{
			name:     "Semaphore with capacity 5",
			capacity: 5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			semaphore := NewSemaphore(tc.capacity)

			// Acquire permits up to the capacity
			for i := 0; i < tc.capacity; i++ {
				semaphore.Acquire()
			}

			// Try to acquire one more permit, should fail
			if semaphore.Try() {
				t.Errorf("Expected Try() to return false when capacity is reached")
			}

			// Release one permit and try again, should succeed
			semaphore.Release()
			if !semaphore.Try() {
				t.Errorf("Expected Try() to return true after releasing a permit")
			}
		})
	}
}
