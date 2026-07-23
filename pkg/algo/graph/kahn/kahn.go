package kahn

import (
	"tsumegolang/pkg/algo/graph/common"
	"tsumegolang/pkg/ds/graph"
)

// Graph is the interface required by TopologicalSort.
type Graph interface {
	IsDirected() bool
	Disconnect(i, j int) error
	GetAllNodes() []int
	GetAllEdges() []graph.Edge
}

// TopologicalSort returns a topological ordering of the nodes in g using Kahn's algorithm.
// NOTE: this mutates g by removing edges; pass a copy if you need to preserve the original.
func TopologicalSort(g Graph) ([]int, error) {
	if !g.IsDirected() {
		return nil, common.ErrNeedDirectedGraph
	}

	set := make(map[int]struct{})
	for _, node := range g.GetAllNodes() {
		set[node] = struct{}{}
	}
	for _, edge := range g.GetAllEdges() {
		delete(set, edge.To.(int))
	}

	sorted := make([]int, 0, len(set))

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
