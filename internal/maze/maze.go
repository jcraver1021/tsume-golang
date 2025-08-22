package maze

import (
	"fmt"

	"tsumegolang/pkg/algorithms/graph/kruskal"
	"tsumegolang/pkg/datastructures/sparsegraph"
)

const (
	defaultWidth  = 4
	defaultHeight = 4
)

type MazeGenerator struct {
	rect           *Rectangle
	recursionLevel int
	filename       string
}

type GeneratorOptions func(*MazeGenerator)

func NewMazeGenerator(width, height int, opts ...GeneratorOptions) *MazeGenerator {
	rect, _ := NewRectangle(width, height, NO_CONNECT)
	mg := &MazeGenerator{
		rect: rect,
		// zero values are fine for the other parameters
	}

	for _, opt := range opts {
		opt(mg)
	}

	return mg
}

func WithFilename(filename string) GeneratorOptions {
	return func(mg *MazeGenerator) {
		mg.filename = filename
	}
}

func WithRecursionLevel(level int) GeneratorOptions {
	return func(mg *MazeGenerator) {
		mg.recursionLevel = level
	}
}

func (mg *MazeGenerator) Generate() error {
	if err := mg.generate(mg.recursionLevel); err != nil {
		return fmt.Errorf("failed to generate maze: %w", err)
	}

	if err := mg.drawMaze(); err != nil {
		return fmt.Errorf("failed to draw maze: %w", err)
	}

	return nil
}

func (mg *MazeGenerator) generate(recursionLevel int) error {
	if recursionLevel > 0 {
		submazes := makeSubmazeMatrix(mg.rect.Width, mg.rect.Height)
		for i := range mg.rect.Width {
			for j := range mg.rect.Height {
				submazes[i][j] = NewMazeGenerator(defaultWidth, defaultHeight)
				err := submazes[i][j].generate(recursionLevel - 1)
				if err != nil {
					return err
				}
			}
		}

		rect, _ := NewRectangle(mg.rect.Width*submazes[0][0].rect.Width, mg.rect.Height*submazes[0][0].rect.Height, NO_CONNECT)
		mg.rect = rect
		mg.join(submazes)
	} else {
		rect, err := NewRectangle(mg.rect.Width, mg.rect.Height, CONNECT_RANDOM)
		if err != nil {
			return err
		}

		mg.rect = rect
	}

	// Filter out the MST to represent the paths)
	baseGraph := mg.rect.Graph
	paths, err := sparsegraph.NewGraph(baseGraph.GetSize(), true)
	if err != nil {
		return err
	}
	err = kruskal.MST(baseGraph, paths)
	if err != nil {
		return err
	}

	mg.rect.Graph = paths
	return nil
}

func makeSubmazeMatrix(width, height int) [][]*MazeGenerator {
	matrix := make([][]*MazeGenerator, height)
	for i := range matrix {
		matrix[i] = make([]*MazeGenerator, width)
	}

	return matrix
}

func (mg *MazeGenerator) join(matrix [][]*MazeGenerator) {
	for i := range matrix {
		for j := range matrix[i] {
			r := matrix[i][j].rect
			rw := r.Width
			rh := r.Height
			iBase := i * rw
			jBase := j * rh
			for ri := range rw {
				for rj := range rh {
					point := r.RectToGraph(ri, rj)
					east := r.RectToGraph(ri+1, rj)
					south := r.RectToGraph(ri, rj+1)
					if e, ok := r.Graph.GetEdge(point, east); ok || (ri+1 == rw && i+1 < len(matrix)) {
						newPoint := mg.rect.RectToGraph(iBase+ri, jBase+rj)
						newEast := mg.rect.RectToGraph(iBase+ri+1, jBase+rj)
						weight := e.Weight
						if !ok {
							weight = getNewWeight(CONNECT_RANDOM)
						}
						mg.rect.Graph.Connect(newPoint, newEast, weight)
					}
					if e, ok := r.Graph.GetEdge(point, south); ok || (rj+1 == rh && j+1 < len(matrix[i])) {
						newPoint := mg.rect.RectToGraph(iBase+ri, jBase+rj)
						newSouth := mg.rect.RectToGraph(iBase+ri, jBase+rj+1)
						weight := e.Weight
						if !ok {
							weight = getNewWeight(CONNECT_RANDOM)
						}
						mg.rect.Graph.Connect(newPoint, newSouth, weight)
					}
				}
			}
		}
	}
}

func (mg *MazeGenerator) drawMaze() error {
	// Draw the maze
	img, err := DrawRectangleMaze(mg.rect)
	if err != nil {
		return err
	}

	// Save the image to a file
	err = WriteImageToFile(&img, mg.filename)
	if err != nil {
		return err
	}

	return nil
}
