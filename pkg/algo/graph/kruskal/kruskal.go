package kruskal

import (
	"tsumegolang/pkg/algo/graph/common"
	"tsumegolang/pkg/ds/disjointset"
	"tsumegolang/pkg/ds/graph"
)

// Graph is the interface required by MST.
type Graph interface {
	GetSize() int
	IsDirected() bool
	Connect(i, j int, w float64) error
	GetAllEdges() []graph.Edge
}

// MST computes the minimum spanning tree of original using Kruskal's algorithm,
// storing the result in mst. mst must be empty. original is not modified.
func MST(original, mst Graph) error {
	if original.IsDirected() {
		return common.ErrNeedUndirectedGraph
	}

	edges := graph.NewEdgeHeap(original.GetAllEdges())

	ds, err := disjointset.NewDisjointSet(disjointset.WithCapacity(original.GetSize()))
	if err != nil {
		return err
	}
	if _, err = ds.AddMany(original.GetSize()); err != nil {
		return err
	}

	for edges.Len() > 0 {
		edge, err := edges.PopEdge()
		if err != nil {
			return err
		}

		rootL, err := ds.Find(edge.From.(int))
		if err != nil {
			return err
		}
		rootR, err := ds.Find(edge.To.(int))
		if err != nil {
			return err
		}

		if rootL != rootR {
			mst.Connect(edge.From.(int), edge.To.(int), edge.Weight)
			if err := ds.Union(rootL, rootR); err != nil {
				return err
			}
		}
	}

	return nil
}
