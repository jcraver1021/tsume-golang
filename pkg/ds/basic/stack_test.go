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
			count := 0
			for i, op := range tc.ops {
				switch op {
				case stackOpPush:
					if count == 0 && !stack.IsEmpty() {
						t.Fatalf("unexpected non-empty stack before first push")
					}
					stack.Push(tc.inputs[i])
					val, ok := stack.Peek()
					if !ok {
						t.Fatalf("unexpected error on peek after push: stack is empty")
					}
					if val != tc.inputs[i] {
						t.Errorf("peek value does not match pushed value: got %v, want %v", val, tc.inputs[i])
					}
					count++
					if stack.Len() != count {
						t.Errorf("stack length mismatch after push: got %v, want %v", stack.Len(), count)
					}
				case stackOpPop:
					val1, ok := stack.Peek()
					if !ok {
						if !tc.wantError {
							t.Fatalf("unexpected error on peek: stack is empty")
						}
						continue
					}
					val2, ok := stack.Pop()
					if !ok {
						if !tc.wantError {
							t.Fatalf("unexpected error on pop: stack is empty")
						}
						continue
					}
					if val1 != val2 {
						t.Errorf("peek and pop values do not match: got %v, want %v", val2, val1)
					}
					count--
					if stack.Len() != count {
						t.Errorf("stack length mismatch after pop: got %v, want %v", stack.Len(), count)
					}
					if count == 0 && !stack.IsEmpty() {
						t.Fatalf("unexpected non-empty stack after popping all elements")
					}
					got = append(got, val2)
				}
			}
			if !tc.wantError && !slices.Equal(got, tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}
