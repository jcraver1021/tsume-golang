package main

import (
	"flag"
	"fmt"

	"tsumegolang/internal/maze"
)

var (
	widthFlag    = flag.Int("width", 20, "Width of the maze")
	heightFlag   = flag.Int("height", 20, "Height of the maze")
	filenameFlag = flag.String("filename", "maze.png", "Output filename for the maze image")
)

func main() {
	flag.Parse()

	err := maze.GenerateMaze(*widthFlag, *heightFlag, *filenameFlag)
	if err != nil {
		fmt.Printf("Error generating maze: %v\n", err)
		return
	}
	fmt.Printf("Maze generated successfully: %s (%dx%d)\n", *filenameFlag, *widthFlag, *heightFlag)
}
