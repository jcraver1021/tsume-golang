package kruskal

import (
	"errors"
)

var (
	ErrDirectedGraph = errors.New("this algorithm requires an undirected graph")
)
