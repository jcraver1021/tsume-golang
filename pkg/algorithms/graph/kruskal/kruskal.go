package kruskal

import (
	"tsumegolang/pkg/datastructures/disjointset"
	"tsumegolang/pkg/datastructures/graph"
)

// Graph interface defines the methods required for a graph to be used with Kruskal's algorithm.
// We do not require a specific graph implementation, but it must support the following methods:
// - GetSize: returns the number of nodes in the graph.
// - ConnectBidirectional: connects two nodes bidirectionally with a given weight.
// - GetAllEdges: returns all edges in the graph as triples [weight, from, to].
type Graph interface {
	GetSize() int
	IsDirected() bool
	Connect(i, j int, w float64) error
	GetAllEdges() []graph.Edge
}

// MST computes the Minimum Spanning Tree (MST) of the original graph using Kruskal's algorithm.
// The result is stored in the mst graph, which should be empty before calling this function.
// The original graph is not modified.
func MST(original, mst Graph) error {
	// Kruskal's algorithm requires an undirected graph.
	if original.IsDirected() {
		return ErrDirectedGraph
	}

	// We will assume that mst is empty.

	// Get all edges from the original graph and create a min-heap.
	edges := graph.NewEdgeHeap(original.GetAllEdges())

	// Create a disjoint set to keep track of connected components.
	ds, err := disjointset.NewDisjointSet(disjointset.WithCapacity(original.GetSize()))
	if err != nil {
		panic(err)
	}
	_, err = ds.AddMany(original.GetSize())
	if err != nil {
		panic(err)
	}

	// Process edges in order of weight.
	for edges.Len() > 0 {
		edge, err := edges.PopEdge()
		if err != nil {
			return err
		}

		// Find the roots of the sets containing each endpoint of the edge.
		rootL, err := ds.Find(edge.From.(int))
		if err != nil {
			panic(err)
		}
		rootR, err := ds.Find(edge.To.(int))
		if err != nil {
			panic(err)
		}

		// If they are in different sets, add the edge to the MST and union the sets.
		if rootL != rootR {
			mst.Connect(edge.From.(int), edge.To.(int), edge.Weight)
			if err := ds.Union(rootL, rootR); err != nil {
				return err
			}
		}
	}

	return nil
}
