package disjointset

import (
	"testing"
)

func TestDisjointSet(t *testing.T) {
	testCases := []struct {
		name          string
		n             int
		union         [][2]int
		wantConnected [][]int
	}{
		{
			name:          "simple union",
			n:             3,
			union:         [][2]int{{0, 1}},
			wantConnected: [][]int{{0, 1}, {2}},
		},
		{
			name:          "multiple unions",
			n:             4,
			union:         [][2]int{{0, 1}, {2, 3}},
			wantConnected: [][]int{{0, 1}, {2, 3}},
		},
		{
			name:          "union all",
			n:             4,
			union:         [][2]int{{0, 1}, {1, 2}, {2, 3}},
			wantConnected: [][]int{{0, 1, 2, 3}},
		},
		{
			name:          "union all with duplicates",
			n:             4,
			union:         [][2]int{{0, 1}, {1, 2}, {2, 3}, {0, 3}},
			wantConnected: [][]int{{0, 1, 2, 3}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds, err := NewDisjointSet()
			if err != nil {
				t.Fatalf("NewDisjointSet() failed: %v", err)
			}
			for i := 0; i < tc.n; i++ {
				ds.Add()
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

func TestDisjointSetCapacity(t *testing.T) {
	testCases := []struct {
		name    string
		n       int
		wantErr error
	}{
		{
			name: "default capacity",
			n:    10,
		},
		{
			name: "high capacity",
			n:    1337,
		},
		{
			name:    "zero capacity",
			n:       0,
			wantErr: ErrInvalidConfiguration,
		},
		{
			name:    "negative capacity",
			n:       -1,
			wantErr: ErrInvalidConfiguration,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds, err := NewDisjointSet(WithCapacity(tc.n))
			if err != nil && err != tc.wantErr {
				t.Fatalf("NewDisjointSet() failed: %v", err)
			} else if err == nil && tc.wantErr != nil {
				t.Fatalf("NewDisjointSet() succeeded, expected error: %v", tc.wantErr)
			} else if err == nil && cap(ds.parents) != tc.n {
				t.Errorf("Expected capacity %d, got %d", tc.n, cap(ds.parents))
			}
		})
	}
}
func TestDisjointSetScaleFactor(t *testing.T) {
	testCases := []struct {
		name    string
		n       int
		wantErr error
	}{
		{
			name: "default scale factor",
			n:    2,
		},
		{
			name: "high scale factor",
			n:    42,
		},
		{
			name:    "zero scale factor",
			n:       0,
			wantErr: ErrInvalidConfiguration,
		},
		{
			name:    "negative scale factor",
			n:       -1,
			wantErr: ErrInvalidConfiguration,
		},
		{
			name:    "scale factor one",
			n:       1,
			wantErr: ErrInvalidConfiguration,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds, err := NewDisjointSet(WithScaleFactor(tc.n))
			if err != nil && err != tc.wantErr {
				t.Fatalf("NewDisjointSet() failed: %v", err)
			} else if err == nil && tc.wantErr != nil {
				t.Fatalf("NewDisjointSet() succeeded, expected error: %v", tc.wantErr)
			}
			if err == nil {
				c := cap(ds.parents)
				// Add one extra over capacity to trigger expansion.
				for i := 0; i <= c; i++ {
					ds.Add()
				}
				if cap(ds.parents) != c*tc.n {
					t.Errorf("Expected capacity %d, got %d", c*tc.n, cap(ds.parents))
				}
			}
		})
	}
}

func TestDisjointSetFindErrors(t *testing.T) {
	testCases := []struct {
		name string
		n    int
		find int
	}{
		{
			name: "find missing element",
			n:    3,
			find: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds, err := NewDisjointSet()
			if err != nil {
				t.Fatalf("NewDisjointSet() failed: %v", err)
			}
			ds.AddMany(tc.n)

			if _, err := ds.Find(tc.find); err != ErrMissingElement {
				t.Errorf("Find(%d) = %v, want %v", tc.find, err, ErrMissingElement)
			}
		})
	}
}

func TestDisjointSetUnionErrors(t *testing.T) {
	testCases := []struct {
		name  string
		n     int
		union [2]int
	}{
		{
			name:  "union missing element",
			n:     3,
			union: [2]int{0, 3},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds, err := NewDisjointSet()
			if err != nil {
				t.Fatalf("NewDisjointSet() failed: %v", err)
			}
			ds.AddMany(tc.n)

			if err := ds.Union(tc.union[0], tc.union[1]); err != ErrMissingElement {
				t.Errorf("Union(%d, %d) = %v, want %v", tc.union[0], tc.union[1], err, ErrMissingElement)
			}
		})
	}
}

func TestDisjointSetAddManyError(t *testing.T) {
	testCases := []struct {
		name string
		n    int
	}{
		{
			name: "add many zero",
			n:    0,
		},
		{
			name: "add many negative",
			n:    -1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ds, err := NewDisjointSet()
			if err != nil {
				t.Fatalf("NewDisjointSet() failed: %v", err)
			}

			if _, err := ds.AddMany(tc.n); err != ErrInvalidRequest {
				t.Errorf("AddMany(%d) = %v, want %v", tc.n, err, ErrInvalidRequest)
			}
		})
	}
}
