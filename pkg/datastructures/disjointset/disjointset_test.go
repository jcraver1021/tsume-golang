package disjointset

import (
	"testing"
)

func TestDisjointSet(t *testing.T) {
	testCases := []struct {
		name          string
		elements      []int
		union         [][2]int
		wantConnected [][]int
	}{
		{
			name:          "simple union",
			elements:      []int{1, 2, 3},
			union:         [][2]int{{1, 2}},
			wantConnected: [][]int{{1, 2}, {3}},
		},
		{
			name:          "multiple unions",
			elements:      []int{1, 2, 3, 4},
			union:         [][2]int{{1, 2}, {3, 4}},
			wantConnected: [][]int{{1, 2}, {3, 4}},
		},
		{
			name:          "union all",
			elements:      []int{1, 2, 3, 4},
			union:         [][2]int{{1, 2}, {2, 3}, {3, 4}},
			wantConnected: [][]int{{1, 2, 3, 4}},
		},
		{
			name:          "union all with duplicates",
			elements:      []int{1, 2, 3, 4},
			union:         [][2]int{{1, 2}, {2, 3}, {3, 4}, {1, 2}},
			wantConnected: [][]int{{1, 2, 3, 4}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := New[int]()
			for _, e := range tc.elements {
				ds.Add(e)
			}

			for _, u := range tc.union {
				if err := ds.Union(u[0], u[1]); err != nil {
					t.Fatalf("Union(%d, %d) failed: %v", u[0], u[1], err)
				}
			}

			for _, want := range tc.wantConnected {
				root, err := ds.Find(want[0])
				if err != nil {
					t.Fatalf("Find(%d) failed: %v", want[0], err)
				}
				for _, e := range want {
					got, err := ds.Find(e)
					if err != nil {
						t.Fatalf("Find(%d) failed: %v", e, err)
					}
					if got != root {
						t.Errorf("Find(%d) = %d, want %d", e, got, root)
					}
				}
			}
		})
	}
}

func TestDisjointSetFindErrors(t *testing.T) {
	testCases := []struct {
		name     string
		elements []int
		find     int
	}{
		{
			name:     "find missing element",
			elements: []int{1, 2, 3},
			find:     4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := New[int]()
			for _, e := range tc.elements {
				ds.Add(e)
			}

			if _, err := ds.Find(tc.find); err != ErrMissingElement {
				t.Errorf("Find(%d) = %v, want %v", tc.find, err, ErrMissingElement)
			}
		})
	}
}

func TestDisjointSetUnionErrors(t *testing.T) {
	testCases := []struct {
		name     string
		elements []int
		union    [2]int
	}{
		{
			name:     "union missing element",
			elements: []int{1, 2, 3},
			union:    [2]int{1, 4},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds := New[int]()
			for _, e := range tc.elements {
				ds.Add(e)
			}

			if err := ds.Union(tc.union[0], tc.union[1]); err != ErrMissingElement {
				t.Errorf("Union(%d, %d) = %v, want %v", tc.union[0], tc.union[1], err, ErrMissingElement)
			}
		})
	}
}
