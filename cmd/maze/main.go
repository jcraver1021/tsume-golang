package main

import (
	"flag"
	"fmt"

	"tsumegolang/internal/maze"
)

var (
	widthFlag     = flag.Int("width", 20, "Width of the maze")
	heightFlag    = flag.Int("height", 20, "Height of the maze")
	filenameFlag  = flag.String("filename", "maze.png", "Output filename for the maze image")
	recursionFlag = flag.Int("recursion", 0, "Recursion level")
)

func doMaze(width, height int, filename string, recursionLevel int) {
	mg := maze.NewMazeGenerator(
		width,
		height,
		maze.WithFilename(filename),
		maze.WithRecursionLevel(recursionLevel),
	)
	err := mg.Generate()
	if err != nil {
		fmt.Printf("Error generating maze: %v\n", err)
		return
	}
}

func main() {
	flag.Parse()

	doMaze(*widthFlag, *heightFlag, *filenameFlag, *recursionFlag)
}
