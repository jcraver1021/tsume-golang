package maze

import (
	"math/rand"

	"tsumegolang/pkg/datastructures/sparsegraph"
)

type EdgeInit int

const (
	NO_CONNECT     EdgeInit = iota // No connection, we just want use the coordinate system.
	CONNECT_CONST                  // Connect every neighbor with weight 1.
	CONNECT_RANDOM                 // Connect every neighbor with weight selected from the half-open interval [0.0, 1.0).
)

const (
	constWeight = 1.0
)

type Rectangle struct {
	Width  int
	Height int
	Graph  *sparsegraph.Graph
}

// NewRectangle generates a rectangular, undirected graph with the specified width and height,
// where each vertex connects to its north, east, south, and west neighbors.
func NewRectangle(width, height int, init EdgeInit) (*Rectangle, error) {
	g, err := sparsegraph.NewGraph(width*height, false)
	if err != nil {
		return nil, err
	}

	r := &Rectangle{
		Width:  width,
		Height: height,
		Graph:  g,
	}

	if init == NO_CONNECT {
		return r, nil
	}

	for y := range height {
		for x := range width {
			current := y*width + x

			// Since the graph is undirected, we only need to consider the east and south neighbors.
			if x < width-1 { // East neighbor
				east := current + 1
				weight := getNewWeight(init)
				err = r.Graph.Connect(current, east, weight)
				if err != nil {
					return nil, err
				}
			}

			if y < height-1 { // South neighbor
				south := current + width
				weight := getNewWeight(init)
				err = r.Graph.Connect(current, south, weight)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return r, nil
}

func getNewWeight(method EdgeInit) float64 {
	switch method {
	case CONNECT_CONST:
		return constWeight
	case CONNECT_RANDOM:
		return rand.Float64()
	}

	return 0
}

func (r *Rectangle) RectToGraph(i, j int) int {
	return i*r.Width + j
}
