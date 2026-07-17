package common

import (
	"errors"
)

var (
	ErrCycleDetected       = errors.New("cycle detected in the graph")
	ErrDisconnectedGraph   = errors.New("the graph is disconnected")
	ErrNeedDirectedGraph   = errors.New("this algorithm requires a directed graph")
	ErrNeedUndirectedGraph = errors.New("this algorithm requires an undirected graph")
)
