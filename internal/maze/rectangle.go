package maze

import (
	"math/rand"

	"tsumegolang/pkg/datastructures/sparsegraph"
)

const (
	DefaultWeight = 1.0 // Default weight for edges if not randomized
)

// Rectangle generates a rectangular, undirected graph with the specified width and height,
// where each vertex connects to its north, east, south, and west neighbors.
// If specified, the weights will be randomized within the half-open interval [0.0, 1.0).
func Rectangle(width, height int, randomize bool) (*sparsegraph.Graph, error) {
	g, err := sparsegraph.NewGraph(width*height, false)
	if err != nil {
		return nil, err
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			current := y*width + x

			// Since the graph is undirected, we only need to consider the east and south neighbors.
			if x < width-1 { // East neighbor
				east := current + 1
				weight := DefaultWeight
				if randomize {
					weight = rand.Float64()
				}
				err = g.Connect(current, east, weight)
				if err != nil {
					return nil, err
				}
			}

			if y < height-1 { // South neighbor
				south := current + width
				weight := 1.0
				if randomize {
					weight = rand.Float64()
				}
				err = g.Connect(current, south, weight)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return g, nil
}
