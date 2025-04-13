package sparsegraph

import (
	"testing"
)

func TestGraph(t *testing.T) {
	testCases := []struct {
		name       string
		nodes      [2]uint64
		w          float64
		checkNodes [2]uint64
		want       float64
		exists     bool
	}{
		{
			name:       "get existing edge",
			nodes:      [2]uint64{1, 2},
			w:          1.5,
			checkNodes: [2]uint64{1, 2},
			want:       1.5,
			exists:     true,
		},
		{
			name:       "get non-existing edge",
			nodes:      [2]uint64{1, 2},
			w:          1.5,
			checkNodes: [2]uint64{2, 3},
			want:       0,
			exists:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g := NewGraph()
			g.Connect(tc.nodes[0], tc.nodes[1], tc.w)
			got, exists := g.GetEdge(tc.checkNodes[0], tc.checkNodes[1])
			if got != tc.want || exists != tc.exists {
				t.Errorf("got (%v, %v), want (%v, %v)", got, exists, tc.want, tc.exists)
			}
		})
	}
}
