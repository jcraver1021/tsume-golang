package sparsegraph

import (
	"errors"
)

var (
	ErrInvalidConfiguration = errors.New("invalid configuration for sparse graph")
	ErrNoSuchNode           = errors.New("no such node in the graph")
	ErrNoSuchEdge           = errors.New("no such edge in the graph")
)
