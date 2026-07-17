package kahn

import (
	"tsumegolang/pkg/algo/graph/common"
	"tsumegolang/pkg/ds/graph"
)

// Graph interface defines the methods required for a graph to be used with Kahn's algorithm.
// We do not require a specific graph implementation, but it must support the following methods:
// - IsDirected: returns true if the graph is directed, false otherwise.
// - Disconnect: removes the directed edge from node i to node j.
// - GetAllNodes: returns a slice of all nodes in the graph.
// - GetAllEdges: returns all edges in the graph as triples [weight, from, to].
type Graph interface {
	IsDirected() bool
	Disconnect(i, j int) error
	GetAllNodes() []int
	GetAllEdges() []graph.Edge
}

// TopologicalSort returns a topological ordering of the nodes in g using Kahn's algorithm.
// NOTE: this mutates g by removing edges; pass a copy if you need to preserve the original.
func TopologicalSort(g Graph) ([]int, error) {
	// Kahn's algorithm requires a directed graph.
	if !g.IsDirected() {
		return nil, common.ErrNeedDirectedGraph
	}

	// Set to keep track of initial nodes with no incoming edges.
	set := make(map[int]struct{})
	for _, node := range g.GetAllNodes() {
		set[node] = struct{}{}
	}
	for _, edge := range g.GetAllEdges() {
		delete(set, edge.To.(int))
	}

	// List to store the sorted order of nodes.
	sorted := make([]int, 0, len(set))

	// While there are unvisited sources, pull them from the set and remove their outgoing edges.
	for len(set) > 0 {
		var source int
		for node := range set {
			source = node
			break
		}
		delete(set, source)
		sorted = append(sorted, source)

		for _, edge := range g.GetAllEdges() {
			if edge.From.(int) == source {
				if err := g.Disconnect(source, edge.To.(int)); err != nil {
					return nil, err
				}
				// If the destination node has no other incoming edges, add it to the set.
				hasIncoming := false
				for _, e := range g.GetAllEdges() {
					if e.To.(int) == edge.To.(int) {
						hasIncoming = true
						break
					}
				}
				if !hasIncoming {
					set[edge.To.(int)] = struct{}{}
				}
			}
		}
	}

	if len(sorted) != len(g.GetAllNodes()) {
		return nil, common.ErrCycleDetected
	}

	return sorted, nil
}
