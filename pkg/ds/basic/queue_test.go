package basic_test

import (
	"slices"
	"testing"

	. "tsumegolang/pkg/ds/basic"
)

type queueOperation int

const (
	queueOpEnqueue queueOperation = iota
	queueOpDequeue
)

func TestQueue(t *testing.T) {
	testCases := []struct {
		name      string
		inputs    []int
		ops       []queueOperation
		want      []int
		wantError bool
	}{
		{
			name:   "enqueue and dequeue",
			inputs: []int{1, 2, 3},
			ops:    []queueOperation{queueOpEnqueue, queueOpEnqueue, queueOpEnqueue, queueOpDequeue, queueOpDequeue, queueOpDequeue},
			want:   []int{1, 2, 3},
		},
		{
			name:      "dequeue from empty queue",
			inputs:    []int{},
			ops:       []queueOperation{queueOpDequeue},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			queue := NewQueue[int]()
			var got []int
			for i, op := range tc.ops {
				switch op {
				case queueOpEnqueue:
					queue.Enqueue(tc.inputs[i])
				case queueOpDequeue:
					val, ok := queue.Dequeue()
					if !ok {
						if !tc.wantError {
							t.Fatalf("unexpected error on dequeue: queue is empty")
						}
						continue
					}
					got = append(got, val)
				}
			}
			if !tc.wantError && !slices.Equal(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
