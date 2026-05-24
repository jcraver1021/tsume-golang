package basic_test

import (
	"slices"
	"testing"

	. "tsumegolang/pkg/ds/basic"
)

type stackOperation int

const (
	stackOpPush stackOperation = iota
	stackOpPop
)

func TestStack(t *testing.T) {
	testCases := []struct {
		name      string
		inputs    []int
		ops       []stackOperation
		want      []int
		wantError bool
	}{
		{
			name:   "push and pop",
			inputs: []int{1, 2, 3},
			ops:    []stackOperation{stackOpPush, stackOpPush, stackOpPush, stackOpPop, stackOpPop, stackOpPop},
			want:   []int{3, 2, 1},
		},
		{
			name:      "pop from empty stack",
			inputs:    []int{},
			ops:       []stackOperation{stackOpPop},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stack := NewStack[int]()
			var got []int
			for i, op := range tc.ops {
				switch op {
				case stackOpPush:
					stack.Push(tc.inputs[i])
				case stackOpPop:
					val, ok := stack.Pop()
					if !ok {
						if !tc.wantError {
							t.Fatalf("unexpected error on pop: stack is empty")
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
