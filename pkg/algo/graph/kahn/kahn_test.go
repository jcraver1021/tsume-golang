package kahn_test

import (
	"testing"

	"tsumegolang/pkg/algo/graph/common"
	. "tsumegolang/pkg/algo/graph/kahn"
	"tsumegolang/pkg/ds/graph"
	"tsumegolang/pkg/ds/graph/sparsegraph"
)

func TestTopologicalSort(t *testing.T) {
	testCases := []struct {
		name        string
		n           int
		connections []graph.Edge
		wantOrders  [][]int // List of valid topological orders (since there can be multiple valid orders).
		wantErr     error
	}{
		{
			name: "simple DAG",
			n:    3,
			connections: []graph.Edge{
				{From: 0, To: 1, Weight: 1.0},
				{From: 1, To: 2, Weight: 1.0},
			},
			wantOrders: [][]int{
				{0, 1, 2},
			},
		},
		{
			name: "DAG with multiple sources",
			n:    4,
			connections: []graph.Edge{
				{From: 0, To: 2, Weight: 1.0},
				{From: 1, To: 2, Weight: 1.0},
				{From: 2, To: 3, Weight: 1.0},
			},
			wantOrders: [][]int{
				{0, 1, 2, 3},
				{1, 0, 2, 3},
			},
		},
		{
			name: "DAG with cycle",
			n:    3,
			connections: []graph.Edge{
				{From: 0, To: 1, Weight: 1.0},
				{From: 1, To: 2, Weight: 1.0},
				{From: 2, To: 0, Weight: 1.0},
			},
			wantErr: common.ErrCycleDetected,
		},
		{
			name: "disconnected DAG",
			n:    4,
			connections: []graph.Edge{
				{From: 0, To: 1, Weight: 1.0},
				{From: 2, To: 3, Weight: 1.0},
			},
			wantOrders: [][]int{
				{0, 1, 2, 3},
				{0, 2, 1, 3},
				{0, 2, 3, 1},
				{2, 0, 1, 3},
				{2, 0, 3, 1},
				{2, 3, 0, 1},
			},
		},
		{
			name: "dense DAG (more edges than nodes)",
			n:    3,
			connections: []graph.Edge{
				{From: 0, To: 1, Weight: 1.0},
				{From: 0, To: 2, Weight: 1.0},
				{From: 1, To: 2, Weight: 1.0},
			},
			wantOrders: [][]int{
				{0, 1, 2},
			},
		},
		{
			name: "bipartite DAG with multiple sources and sinks",
			n:    4,
			connections: []graph.Edge{
				{From: 0, To: 2, Weight: 1.0},
				{From: 0, To: 3, Weight: 1.0},
				{From: 1, To: 2, Weight: 1.0},
				{From: 1, To: 3, Weight: 1.0},
			},
			wantOrders: [][]int{
				{0, 1, 2, 3},
				{0, 1, 3, 2},
				{1, 0, 2, 3},
				{1, 0, 3, 2},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := sparsegraph.NewGraph(tc.n, true)
			if err != nil {
				t.Fatalf("failed to create graph: %v", err)
			}
			for _, conn := range tc.connections {
				g.Connect(conn.From.(int), conn.To.(int), conn.Weight)
			}

			gotOrder, gotErr := TopologicalSort(g)
			if gotErr != nil {
				if tc.wantErr == nil || gotErr.Error() != tc.wantErr.Error() {
					t.Fatalf("unexpected error: got %v, want %v", gotErr, tc.wantErr)
				}
				return
			}

			if tc.wantErr != nil {
				t.Fatalf("expected error %v but got none", tc.wantErr)
			}

			// Check if the obtained order is one of the valid topological orders.
			valid := false
			for _, order := range tc.wantOrders {
				if equal(gotOrder, order) {
					valid = true
					break
				}
			}
			if !valid {
				t.Fatalf("got order %v, which is not a valid topological order", gotOrder)
			}
		})
	}
}

// Helper function to check if two slices are equal.
func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
