package maze

import (
	"math/rand"

	"tsumegolang/pkg/ds/graph/sparsegraph"
)

type EdgeInit int

const (
	NoConnect     EdgeInit = iota // no edges; use the coordinate system only
	ConnectConst                  // connect every neighbor with weight 1
	ConnectRandom                 // connect every neighbor with weight in [0.0, 1.0)
)

const (
	DefaultWeight = 1.0
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

	if init == NoConnect {
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
	case ConnectConst:
		return DefaultWeight
	case ConnectRandom:
		return rand.Float64()
	}

	return 0
}

func (r *Rectangle) RectToGraph(i, j int) int {
	return i*r.Width + j
}
