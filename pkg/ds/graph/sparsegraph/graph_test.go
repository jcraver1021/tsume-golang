package sparsegraph_test

import (
	"testing"

	"tsumegolang/pkg/ds/graph"
	. "tsumegolang/pkg/ds/graph/sparsegraph"
)

func TestCopy(t *testing.T) {
	testCases := []struct {
		name     string
		directed bool
		edges    []graph.Edge
	}{
		{
			name:     "empty graph",
			directed: true,
			edges:    []graph.Edge{},
		},
		{
			name:     "directed graph",
			directed: true,
			edges: []graph.Edge{
				{From: 0, To: 1, Weight: 1.0},
				{From: 1, To: 2, Weight: 2.0},
			},
		},
		{
			name:     "undirected graph",
			directed: false,
			edges: []graph.Edge{
				{From: 0, To: 1, Weight: 1.0},
				{From: 1, To: 2, Weight: 2.0},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGraph(3, tc.directed)
			if err != nil {
				t.Fatalf("NewGraph() failed: %v", err)
			}
			for _, e := range tc.edges {
				if err := g.Connect(e.From.(int), e.To.(int), e.Weight); err != nil {
					t.Fatalf("Connect() failed: %v", err)
				}
			}

			cp := g.Copy()

			// Copy has the same edges.
			want := graph.NewEdgeHeap(g.GetAllEdges())
			got := graph.NewEdgeHeap(cp.GetAllEdges())
			if len(got) != len(want) {
				t.Fatalf("Copy() edge count = %d; want %d", len(got), len(want))
			}
			for i := range got {
				if !got[i].Equals(want[i]) {
					t.Errorf("Copy() edge %d = %v; want %v", i, got[i], want[i])
				}
			}

			// Mutations to the copy do not affect the original.
			if len(tc.edges) > 0 {
				e := tc.edges[0]
				cp.Disconnect(e.From.(int), e.To.(int))
				if _, ok := g.GetEdge(e.From.(int), e.To.(int)); !ok {
					t.Errorf("Disconnect on copy removed edge from original")
				}
			}
		})
	}
}

func TestGetSize(t *testing.T) {
	testCases := []struct {
		name string
		n    int
	}{
		{name: "single node", n: 1},
		{name: "three nodes", n: 3},
		{name: "ten nodes", n: 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGraph(tc.n, true)
			if err != nil {
				t.Fatalf("NewGraph() failed: %v", err)
			}
			if got := g.GetSize(); got != tc.n {
				t.Errorf("GetSize() = %d; want %d", got, tc.n)
			}
		})
	}
}

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

func TestDisconnect(t *testing.T) {
	testCases := []struct {
		name     string
		directed bool
		setup    []graph.Edge
		i, j     int
		wantErr  error
	}{
		{
			name:     "remove existing directed edge",
			directed: true,
			setup:    []graph.Edge{{From: 0, To: 1, Weight: 1.0}},
			i:        0,
			j:        1,
			wantErr:  nil,
		},
		{
			name:     "remove existing undirected edge",
			directed: false,
			setup:    []graph.Edge{{From: 0, To: 1, Weight: 1.0}},
			i:        1,
			j:        0,
			wantErr:  nil,
		},
		{
			name:     "edge does not exist",
			directed: true,
			setup:    []graph.Edge{},
			i:        0,
			j:        1,
			wantErr:  ErrNoSuchEdge,
		},
		{
			name:     "invalid node index",
			directed: true,
			setup:    []graph.Edge{},
			i:        -1,
			j:        1,
			wantErr:  ErrNoSuchNode,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGraph(3, tc.directed)
			if err != nil {
				t.Fatalf("NewGraph() failed: %v", err)
			}
			for _, e := range tc.setup {
				if err := g.Connect(e.From.(int), e.To.(int), e.Weight); err != nil {
					t.Fatalf("Connect() failed: %v", err)
				}
			}

			err = g.Disconnect(tc.i, tc.j)
			if err != tc.wantErr {
				t.Errorf("Disconnect(%d, %d) = %v; want %v", tc.i, tc.j, err, tc.wantErr)
			}

			if tc.wantErr == nil {
				if _, ok := g.GetEdge(tc.i, tc.j); ok {
					t.Errorf("edge (%d, %d) still exists after Disconnect", tc.i, tc.j)
				}
			}
		})
	}
}

func TestGetAllNodes(t *testing.T) {
	testCases := []struct {
		name string
		n    int
	}{
		{name: "single node", n: 1},
		{name: "three nodes", n: 3},
		{name: "ten nodes", n: 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGraph(tc.n, true)
			if err != nil {
				t.Fatalf("NewGraph() failed: %v", err)
			}

			nodes := g.GetAllNodes()
			if len(nodes) != tc.n {
				t.Fatalf("GetAllNodes() length = %d; want %d", len(nodes), tc.n)
			}
			seen := make(map[int]struct{})
			for _, node := range nodes {
				if node < 0 || node >= tc.n {
					t.Errorf("GetAllNodes() returned out-of-range node %d", node)
				}
				seen[node] = struct{}{}
			}
			if len(seen) != tc.n {
				t.Errorf("GetAllNodes() returned duplicate nodes")
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

			for _, e := range tc.edges {
				if err := g.Connect(e.From.(int), e.To.(int), e.Weight); err != nil {
					t.Fatalf("Connect(%d, %d, %f) failed: %v", e.From, e.To, e.Weight, err)
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
			for i := range got {
				if !got[i].Equals(want[i]) {
					t.Errorf("GetAllEdges() edge %d = %v; want %v", i, got[i], want[i])
				}
			}
		})
	}
}
