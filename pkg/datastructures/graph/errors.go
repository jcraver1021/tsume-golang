package graph

import (
	"errors"
)

var (
	ErrEdgeHeapEmpty = errors.New("edge heap is empty")
	ErrEdgeHeapType  = errors.New("edge heap type mismatch")
)
