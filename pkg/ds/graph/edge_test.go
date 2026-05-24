package graph

import (
	"testing"
)

func TestEdgeEquals(t *testing.T) {
	testCases := []struct {
		name   string
		e1, e2 Edge
		want   bool
	}{
		{
			name: "equal edges",
			e1:   Edge{From: 1, To: 2, Weight: 3.0},
			e2:   Edge{From: 1, To: 2, Weight: 3.0},
			want: true,
		},
		{
			name: "different from node",
			e1:   Edge{From: 1, To: 2, Weight: 3.0},
			e2:   Edge{From: 2, To: 2, Weight: 3.0},
			want: false,
		},
		{
			name: "different to node",
			e1:   Edge{From: 1, To: 2, Weight: 3.0},
			e2:   Edge{From: 1, To: 3, Weight: 3.0},
			want: false,
		},
		{
			name: "different weight",
			e1:   Edge{From: 1, To: 2, Weight: 3.0},
			e2:   Edge{From: 1, To: 2, Weight: 4.0},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.e1.Equals(tc.e2); got != tc.want {
				t.Errorf("Edge.Equals() = %v; want %v", got, tc.want)
			}
		})
	}
}

func TestEdgeHeap(t *testing.T) {
	testCases := []struct {
		name  string
		edges []Edge
		want  []Edge
	}{
		{
			name:  "empty heap",
			edges: []Edge{},
			want:  []Edge{},
		},
		{
			name: "single edge",
			edges: []Edge{
				{From: 1, To: 2, Weight: 3.0},
			},
			want: []Edge{
				{From: 1, To: 2, Weight: 3.0},
			},
		},
		{
			name: "multiple edges",
			edges: []Edge{
				{From: 1, To: 2, Weight: 3.0},
				{From: 2, To: 3, Weight: 1.0},
				{From: 3, To: 4, Weight: 2.0},
			},
			want: []Edge{
				{From: 2, To: 3, Weight: 1.0},
				{From: 3, To: 4, Weight: 2.0},
				{From: 1, To: 2, Weight: 3.0},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewEdgeHeap(tc.edges)
			if len(h) != len(tc.want) {
				t.Fatalf("NewEdgeHeap() length = %d; want %d", len(h), len(tc.want))
			}
			for _, want := range tc.want {
				got, err := h.PopEdge()
				if err != nil {
					t.Fatalf("NewEdgeHeap().Pop() did not return Edge type: %v", err)
				}
				if !got.Equals(want) {
					t.Errorf("NewEdgeHeap().Pop() = %v; want %v", got, want)
				}
			}
		})
	}
}
