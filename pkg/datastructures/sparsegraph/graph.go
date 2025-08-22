// Package sparsegraph implements a sparse graph data structure. The number of
// nodes is immutable after creation.

package sparsegraph

import (
	"tsumegolang/pkg/datastructures/graph"
)

type Graph struct {
	directed bool              // Indicates if the graph is directed or not.
	nodes    []map[int]float64 // Maps node index to its adjacency list, with the value being the weight of the edge.
}

// NewGraph creates a new sparse graph with the specified number of nodes.
// Returns an error if the number of nodes is less than or equal to zero.
func NewGraph(n int, directed bool) (*Graph, error) {
	if n <= 0 {
		return nil, ErrInvalidConfiguration
	}

	return &Graph{
		directed: directed,
		nodes:    make([]map[int]float64, n),
	}, nil
}

// checkIdx checks if the given index is valid for the graph.
func (g *Graph) checkIdx(i int) error {
	if i < 0 || i >= len(g.nodes) {
		return ErrNoSuchNode
	}

	return nil
}

// IsDirected returns true if the graph is directed, false if it is undirected.
func (g *Graph) IsDirected() bool {
	return g.directed
}

// GetSize returns the number of nodes in the graph.
// This is immutable after creation.
func (g *Graph) GetSize() int {
	return len(g.nodes)
}

// Connect creates a directed edge from node i to node j with weight w.
// Returns an error if either node index is invalid.
func (g *Graph) Connect(i, j int, w float64) error {
	if err := g.checkIdx(i); err != nil {
		return err
	}
	if err := g.checkIdx(j); err != nil {
		return err
	}

	// Special consideration for undirected graphs.
	if !g.directed && i > j {
		// Arbitrarily store the edge in the lower index node's adjacency list.
		i, j = j, i
	}

	if g.nodes[i] == nil {
		g.nodes[i] = make(map[int]float64)
	}

	g.nodes[i][j] = w

	return nil
}

// GetEdge retrieves the edge from node i to node j.
// Returns the edge and a boolean indicating whether the edge exists.
func (g *Graph) GetEdge(i, j int) (graph.Edge, bool) {
	if err := g.checkIdx(i); err != nil {
		return graph.Edge{}, false
	}
	if err := g.checkIdx(j); err != nil {
		return graph.Edge{}, false
	}

	// Special consideration for undirected graphs.
	if !g.directed && i > j {
		// If the graph is undirected, check the lower-index node's adjacency list.
		i, j = j, i
	}

	if g.nodes[i] == nil {
		return graph.Edge{}, false
	}

	w, exists := g.nodes[i][j]
	return graph.Edge{From: i, To: j, Weight: w}, exists
}

// GetAllEdges returns all edges in the graph as a slice of triples [weight, from, to].
func (g *Graph) GetAllEdges() []graph.Edge {
	edges := make([]graph.Edge, 0)
	for i, adj := range g.nodes {
		if adj == nil {
			continue
		}
		for j, w := range adj {
			edges = append(edges, graph.Edge{From: i, To: j, Weight: w})
		}
	}
	return edges
}
