package kruskal

import (
	"testing"

	"tsumegolang/pkg/datastructures/graph"
	"tsumegolang/pkg/datastructures/sparsegraph"
)

func TestMST(t *testing.T) {
	testCases := []struct {
		name        string
		n           int
		connections []graph.Edge
		wantMST     []graph.Edge
	}{
		{
			name: "simple triangle",
			n:    3,
			connections: []graph.Edge{
				{From: 0, To: 1, Weight: 1.0},
				{From: 1, To: 2, Weight: 1.1},
				{From: 0, To: 2, Weight: 1.5},
			},
			wantMST: []graph.Edge{
				{From: 0, To: 1, Weight: 1.0},
				{From: 1, To: 2, Weight: 1.1},
			},
		},
		{
			name: "fully connected square",
			n:    4,
			connections: []graph.Edge{
				{From: 0, To: 1, Weight: 1.1},
				{From: 1, To: 2, Weight: 2.2},
				{From: 2, To: 3, Weight: 3.2},
				{From: 0, To: 3, Weight: 1.3},
				{From: 0, To: 2, Weight: 1.2},
				{From: 1, To: 3, Weight: 2.3},
			},
			wantMST: []graph.Edge{
				{From: 0, To: 1, Weight: 1.1},
				{From: 0, To: 2, Weight: 1.2},
				{From: 0, To: 3, Weight: 1.3},
			},
		},
		{
			name: "two disconnected squares",
			n:    8,
			connections: []graph.Edge{
				{From: 0, To: 1, Weight: 1.1},
				{From: 1, To: 2, Weight: 2.2},
				{From: 2, To: 3, Weight: 3.2},
				{From: 0, To: 3, Weight: 1.3},
				{From: 0, To: 2, Weight: 1.2},
				{From: 1, To: 3, Weight: 2.3},
				{From: 4, To: 5, Weight: 5.4},
				{From: 5, To: 6, Weight: 5.6},
				{From: 6, To: 7, Weight: 3.6},
				{From: 4, To: 7, Weight: 3.7},
				{From: 4, To: 6, Weight: 4.6},
				{From: 5, To: 7, Weight: 3.5},
			},
			wantMST: []graph.Edge{
				{From: 0, To: 1, Weight: 1.1},
				{From: 0, To: 2, Weight: 1.2},
				{From: 0, To: 3, Weight: 1.3},
				{From: 6, To: 7, Weight: 3.6},
				{From: 4, To: 7, Weight: 3.7},
				{From: 5, To: 7, Weight: 3.5},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create the original graph.
			original, err := sparsegraph.NewGraph(tc.n, false)
			if err != nil {
				t.Fatalf("NewGraph(%d) failed: %v", tc.n, err)
			}
			for _, edge := range tc.connections {
				if err := original.Connect(edge.From.(int), edge.To.(int), edge.Weight); err != nil {
					t.Fatalf("Connect(%d, %d, %f) failed: %v", edge.From, edge.To, edge.Weight, err)
				}
			}

			// Create an empty MST graph.
			mst, err := sparsegraph.NewGraph(tc.n, true)
			if err != nil {
				t.Fatalf("NewGraph(%d) for MST failed: %v", tc.n, err)
			}

			// Compute the MST.
			MST(original, mst)

			// Check if the MST matches the expected result.
			gotMST := mst.GetAllEdges()
			if len(gotMST) != len(tc.wantMST) {
				t.Errorf("MST length = %d; want %d", len(gotMST), len(tc.wantMST))
				return
			}

			wantMSTHeap := graph.NewEdgeHeap(tc.wantMST)
			gotMSTHeap := graph.NewEdgeHeap(gotMST)
			for wantMSTHeap.Len() > 0 {
				want, err := wantMSTHeap.PopEdge()
				if err != nil {
					t.Fatalf("PopEdge from wantMSTHeap failed: %v", err)
				}
				got, err := gotMSTHeap.PopEdge()
				if err != nil {
					t.Fatalf("PopEdge from gotMSTHeap failed: %v", err)
				}
				if !got.Equals(want) {
					t.Errorf("MST edge mismatch: got %v, want %v", got, want)
				}
			}
		})
	}
}

func TestMSTError(t *testing.T) {
	testCases := []struct {
		name        string
		n           int
		connections []graph.Edge
		wantErr     error
	}{
		{
			name: "directed graph",
			n:    3,
			connections: []graph.Edge{
				{From: 0, To: 1, Weight: 1.0},
				{From: 1, To: 2, Weight: 1.1},
				{From: 2, To: 0, Weight: 1.5},
			},
			wantErr: ErrDirectedGraph,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a directed graph.
			directedGraph, err := sparsegraph.NewGraph(tc.n, true)
			if err != nil {
				t.Fatalf("NewGraph(%d) failed: %v", tc.n, err)
			}
			for _, edge := range tc.connections {
				if err := directedGraph.Connect(edge.From.(int), edge.To.(int), edge.Weight); err != nil {
					t.Fatalf("Connect(%d, %d, %f) failed: %v", edge.From, edge.To, edge.Weight, err)
				}
			}

			mst, err := sparsegraph.NewGraph(tc.n, true)
			if err != nil {
				t.Fatalf("NewGraph(%d) for MST failed: %v", tc.n, err)
			}

			err = MST(directedGraph, mst)
			if err != tc.wantErr {
				t.Errorf("MST() error = %v; want %v", err, tc.wantErr)
			}
		})
	}
}
