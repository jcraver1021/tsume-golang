package basic_test

import (
	"testing"

	. "tsumegolang/pkg/ds/basic"
)

type dequeOperation int

const (
	pushFront dequeOperation = iota
	pushBack
	popFront
	popBack
)

func TestDeque(t *testing.T) {
	testCases := []struct {
		name         string
		operations   []dequeOperation
		pushValues   []int
		expectedPops []int
		wantError    bool
	}{
		{
			name:         "push and pop from both ends",
			operations:   []dequeOperation{pushFront, pushBack, pushFront, pushBack, popFront, popBack, popFront, popBack},
			pushValues:   []int{1, 2, 3, 4},
			expectedPops: []int{3, 4, 1, 2},
		},
		{
			name:         "pop from empty deque",
			operations:   []dequeOperation{popFront},
			pushValues:   []int{},
			expectedPops: []int{},
			wantError:    true,
		},
		{
			name: "multiple resizes with mixed operations",
			// Initial capacity: 8, first resize at 8->16, second resize at 16->32
			// Push 20 elements to trigger 2 resizes, then pop them all to verify order
			operations: []dequeOperation{
				// First 10: alternate front/back
				pushFront, pushBack, pushFront, pushBack, pushFront,
				pushBack, pushFront, pushBack, pushFront, pushBack,
				// Next 10: all to back (triggers resizes)
				pushBack, pushBack, pushBack, pushBack, pushBack,
				pushBack, pushBack, pushBack, pushBack, pushBack,
				// Pop all 20 from front
				popFront, popFront, popFront, popFront, popFront,
				popFront, popFront, popFront, popFront, popFront,
				popFront, popFront, popFront, popFront, popFront,
				popFront, popFront, popFront, popFront, popFront,
			},
			pushValues: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
			// After operations, deque order from front to back should be:
			// [9, 7, 5, 3, 1, 2, 4, 6, 8, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20]
			expectedPops: []int{9, 7, 5, 3, 1, 2, 4, 6, 8, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dq := NewDeque[int]()
			pushIdx := 0
			popIdx := 0

			for _, op := range tc.operations {
				switch op {
				case pushFront:
					dq.PushFront(tc.pushValues[pushIdx])
					pushIdx++
				case pushBack:
					dq.PushBack(tc.pushValues[pushIdx])
					pushIdx++
				case popFront:
					val, ok := dq.PopFront()
					if tc.wantError {
						if ok {
							t.Fatalf("expected error: deque should be empty")
						}
					} else {
						if !ok {
							t.Fatalf("unexpected error: deque is empty")
						}
						if val != tc.expectedPops[popIdx] {
							t.Errorf("popFront: expected %v, got %v", tc.expectedPops[popIdx], val)
						}
						popIdx++
					}
				case popBack:
					val, ok := dq.PopBack()
					if tc.wantError {
						if ok {
							t.Fatalf("expected error: deque should be empty")
						}
					} else {
						if !ok {
							t.Fatalf("unexpected error: deque is empty")
						}
						if val != tc.expectedPops[popIdx] {
							t.Errorf("popBack: expected %v, got %v", tc.expectedPops[popIdx], val)
						}
						popIdx++
					}
				}
			}
		})
	}
}

type dequeState struct {
	wantLen   int
	wantFront int
	wantBack  int
	frontOk   bool // whether Front() should return ok=true
	backOk    bool // whether Back() should return ok=true
}

func TestDequeStateVerification(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func(*Deque[int]) // operations to perform
		wantStates []dequeState      // expected states after each setup step
	}{
		{
			name: "empty deque",
			setup: func(dq *Deque[int]) {
				// no-op, test empty state
			},
			wantStates: []dequeState{
				{wantLen: 0, frontOk: false, backOk: false},
			},
		},
		{
			name: "single element",
			setup: func(dq *Deque[int]) {
				dq.PushBack(42)
			},
			wantStates: []dequeState{
				{wantLen: 1, wantFront: 42, wantBack: 42, frontOk: true, backOk: true},
			},
		},
		{
			name: "front and back track correctly",
			setup: func(dq *Deque[int]) {
				dq.PushBack(1)
				dq.PushBack(2)
				dq.PushFront(0)
				dq.PushBack(3)
			},
			wantStates: []dequeState{
				{wantLen: 4, wantFront: 0, wantBack: 3, frontOk: true, backOk: true},
			},
		},
		{
			name: "len updates correctly",
			setup: func(dq *Deque[int]) {
				dq.PushBack(1)
				dq.PushBack(2)
				dq.PushFront(0)
				dq.PopFront()
				dq.PopBack()
				dq.PopFront()
			},
			wantStates: []dequeState{
				{wantLen: 0, frontOk: false, backOk: false},
			},
		},
		{
			name: "empty then push again",
			setup: func(dq *Deque[int]) {
				dq.PushBack(1)
				dq.PushBack(2)
				dq.PopFront()
				dq.PopFront()
				dq.PushBack(10)
				dq.PushFront(9)
			},
			wantStates: []dequeState{
				{wantLen: 2, wantFront: 9, wantBack: 10, frontOk: true, backOk: true},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dq := NewDeque[int]()
			tc.setup(dq)

			for i, want := range tc.wantStates {
				// Check length
				if got := dq.Len(); got != want.wantLen {
					t.Errorf("state %d: Len() = %d, want %d", i, got, want.wantLen)
				}

				// Check Front()
				front, ok := dq.Front()
				if ok != want.frontOk {
					t.Errorf("state %d: Front() ok = %v, want %v", i, ok, want.frontOk)
				}
				if want.frontOk && front != want.wantFront {
					t.Errorf("state %d: Front() = %d, want %d", i, front, want.wantFront)
				}

				// Check Back()
				back, ok := dq.Back()
				if ok != want.backOk {
					t.Errorf("state %d: Back() ok = %v, want %v", i, ok, want.backOk)
				}
				if want.backOk && back != want.wantBack {
					t.Errorf("state %d: Back() = %d, want %d", i, back, want.wantBack)
				}
			}
		})
	}
}

func TestDequeEdgeCases(t *testing.T) {
	testCases := []struct {
		name string
		test func(*testing.T)
	}{
		{
			name: "wraparound without resize",
			test: func(t *testing.T) {
				dq := NewDeque[int]()

				// Push 5 elements
				for i := range 5 {
					dq.PushBack(i)
				}

				// Pop 5 elements (moves head/tail around the ring)
				for i := range 5 {
					val, ok := dq.PopFront()
					if !ok || val != i {
						t.Errorf("PopFront expected %d, got %v", i, val)
					}
				}

				// Now deque is empty but pointers have moved
				// Push 5 more - should wrap around
				for i := range 5 {
					dq.PushBack(i + 10)
				}

				// Verify length and contents
				if dq.Len() != 5 {
					t.Errorf("After wraparound, Len() expected 5, got %d", dq.Len())
				}

				for i := range 5 {
					val, ok := dq.PopFront()
					if !ok || val != i+10 {
						t.Errorf("After wraparound, PopFront expected %d, got %v", i+10, val)
					}
				}
			},
		},
		{
			name: "fill to exactly full then resize",
			test: func(t *testing.T) {
				dq := NewDeque[int]()

				// Initial capacity is 8, usable capacity is 7 (one slot sacrificed)
				// Fill to exactly 7 elements
				for i := range 7 {
					dq.PushBack(i)
				}

				if dq.Len() != 7 {
					t.Errorf("Before resize, Len() expected 7, got %d", dq.Len())
				}

				// Next push should trigger resize
				dq.PushBack(7)

				if dq.Len() != 8 {
					t.Errorf("After resize, Len() expected 8, got %d", dq.Len())
				}

				// Verify all elements are intact
				for i := range 8 {
					val, ok := dq.PopFront()
					if !ok || val != i {
						t.Errorf("After resize, PopFront expected %d, got %v", i, val)
					}
				}
			},
		},
		{
			name: "alternating push and pop",
			test: func(t *testing.T) {
				dq := NewDeque[int]()

				// Alternate push and pop to stress pointer movement
				// Pattern: Push(i), then if odd, pop
				// Result: final deque contains [10, 11, 12, ..., 19]
				for i := range 20 {
					dq.PushBack(i)
					if i%2 == 1 {
						dq.PopFront()
					}
				}

				// Should have 10 elements left
				if dq.Len() != 10 {
					t.Errorf("After alternating push/pop, Len() expected 10, got %d", dq.Len())
				}

				// Verify contents (should be 10 through 19)
				for i := range 10 {
					val, ok := dq.PopFront()
					expected := i + 10
					if !ok || val != expected {
						t.Errorf("Expected %d, got %v", expected, val)
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t)
		})
	}
}
