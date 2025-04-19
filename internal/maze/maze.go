package maze

import (
	"tsumegolang/pkg/algorithms/graph/kruskal"
	"tsumegolang/pkg/datastructures/sparsegraph"
)

// Generate a rectangular maze with the given settings
func GenerateMaze(width, height int, filename string) error {
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
	img, err := DrawRectangleMaze(paths, width, height)
	if err != nil {
		return err
	}

	// Save the image to a file
	err = WriteImageToFile(&img, filename)
	if err != nil {
		return err
	}

	return nil
}
