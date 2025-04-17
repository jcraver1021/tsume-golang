package maze

import (
	"tsumegolang/pkg/algorithms/graph/kruskal"
	"tsumegolang/pkg/datastructures/sparsegraph"
)

// Generate a rectangular maze with the given settings
// TODO: add image output
func GenerateMaze(width, height int) error {
	// Generate the starting graph
	rect, err := Rectangle(width, height, true)
	if err != nil {
		return err
	}

	// Filter out the MST to represent the paths
	paths, err := sparsegraph.NewGraph(rect.GetSize(), true)
	if err != nil {
		return err
	}
	err = kruskal.MST(rect, paths)
	if err != nil {
		return err
	}

	// Draw the maze

	return nil
}
