package graph

import (
	"container/heap"
)

type Edge struct {
	From   any     // The starting node of the edge.
	To     any     // The ending node of the edge.
	Weight float64 // The weight of the edge.
}

func (e Edge) Equals(other Edge) bool {
	return e.From == other.From && e.To == other.To && e.Weight == other.Weight
}

type EdgeHeap []Edge

func (h EdgeHeap) Len() int           { return len(h) }
func (h EdgeHeap) Less(i, j int) bool { return h[i].Weight < h[j].Weight }
func (h EdgeHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *EdgeHeap) Push(x any) {
	*h = append(*h, x.(Edge))
}
func (h *EdgeHeap) Pop() any {
	old := *h
	n := len(old)
	e := old[n-1]
	*h = old[0 : n-1]
	return e
}

func NewEdgeHeap(edges []Edge) EdgeHeap {
	h := EdgeHeap(edges)
	heap.Init(&h)
	return h
}

func (h *EdgeHeap) PopEdge() (Edge, error) {
	if len(*h) == 0 {
		return Edge{}, ErrEdgeHeapEmpty
	}

	edge, ok := heap.Pop(h).(Edge)
	if !ok {
		return Edge{}, ErrEdgeHeapType
	}

	return edge, nil
}
