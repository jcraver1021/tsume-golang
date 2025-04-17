package sparsegraph

import (
	"testing"
	"tsumegolang/pkg/datastructures/graph"
)

func TestConnect(t *testing.T) {
	testCases := []struct {
		name    string
		n       int
		i, j    int
		w       float64
		wantErr error
	}{
		{
			name:    "valid connection",
			n:       3,
			i:       0,
			j:       1,
			w:       1.5,
			wantErr: nil,
		},
		{
			name:    "invalid node index",
			n:       3,
			i:       -1,
			j:       1,
			w:       1.5,
			wantErr: ErrNoSuchNode,
		},
		{
			name:    "invalid node index 2",
			n:       3,
			i:       0,
			j:       3,
			w:       1.5,
			wantErr: ErrNoSuchNode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGraph(tc.n, true)
			if err != nil {
				t.Fatalf("NewGraph() failed: %v", err)
			}

			err = g.Connect(tc.i, tc.j, tc.w)
			if err != tc.wantErr {
				t.Errorf("Connect(%d, %d, %f) = %v; want error? %v", tc.i, tc.j, tc.w, err, tc.wantErr)
			}

			e, ok := g.GetEdge(tc.i, tc.j)
			if err == nil && (!ok || e.Weight != tc.w) {
				t.Errorf("GetEdge(%d, %d) = %f; want %f", tc.i, tc.j, e, tc.w)
			}

			_, ok = g.GetEdge(tc.j, tc.i)
			if err == nil && ok {
				t.Errorf("GetEdge(%d, %d) should not exist after Connect(%d, %d, %f)", tc.j, tc.i, tc.i, tc.j, tc.w)
			}
		})
	}
}

func TestUndirectedConnect(t *testing.T) {
	testCases := []struct {
		name    string
		n       int
		i, j    int
		w       float64
		wantErr error
	}{
		{
			name:    "valid undirected connection",
			n:       3,
			i:       0,
			j:       1,
			w:       1.5,
			wantErr: nil,
		},
		{
			name:    "invalid node index",
			n:       3,
			i:       -1,
			j:       1,
			w:       1.5,
			wantErr: ErrNoSuchNode,
		},
		{
			name:    "invalid node index 2",
			n:       3,
			i:       0,
			j:       3,
			w:       1.5,
			wantErr: ErrNoSuchNode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGraph(tc.n, false)
			if err != nil {
				t.Fatalf("NewGraph() failed: %v", err)
			}

			err = g.Connect(tc.i, tc.j, tc.w)
			if err != tc.wantErr {
				t.Errorf("Connect(%d, %d, %f) = %v; want error? %v", tc.i, tc.j, tc.w, err, tc.wantErr)
			}

			e, ok := g.GetEdge(tc.i, tc.j)
			if err == nil && (!ok || e.Weight != tc.w) {
				t.Errorf("GetEdge(%d, %d) = %f; want %f", tc.i, tc.j, e, tc.w)
			}

			e2, ok2 := g.GetEdge(tc.j, tc.i)
			if err == nil && (!ok2 || e2.Weight != tc.w) {
				t.Errorf("GetEdge(%d, %d) = %f; want %f", tc.j, tc.i, e2, tc.w)
			}
		})
	}
}

func TestGetAllEdges(t *testing.T) {
	testCases := []struct {
		name    string
		n       int
		edges   []graph.Edge
		wantErr error
	}{
		{
			name:    "no edges",
			n:       3,
			edges:   []graph.Edge{},
			wantErr: nil,
		},
		{
			name: "some edges",
			n:    3,
			edges: []graph.Edge{
				{Weight: 1.0, From: 0, To: 1},
				{Weight: 2.0, From: 1, To: 2},
				{Weight: 3.0, From: 2, To: 0},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGraph(tc.n, true)
			if err != nil {
				t.Fatalf("NewGraph() failed: %v", err)
			}

			for _, edge := range tc.edges {
				if err := g.Connect(edge.From.(int), edge.To.(int), edge.Weight); err != nil {
					t.Fatalf("Connect(%d, %d, %f) failed: %v", edge.From, edge.To, edge.Weight, err)
				}
			}

			gotEdges := g.GetAllEdges()
			if len(gotEdges) != len(tc.edges) {
				t.Errorf("GetAllEdges() length = %d; want %d", len(gotEdges), len(tc.edges))
				return
			}

			// Check all edges in ascending order by weight.
			want := graph.NewEdgeHeap(tc.edges)
			got := graph.NewEdgeHeap(gotEdges)
			for i := 0; i < len(got); i++ {
				if !got[i].Equals(want[i]) {
					t.Errorf("GetAllEdges() edge %d = %v; want %v", i, got[i], want[i])
				}
			}
		})
	}
}
