package main

import (
	"fmt"

	"tsumegolang/internal/maze"
)

func main() {
	width := 20
	height := 20
	err := maze.GenerateMaze(width, height, "maze.png")
	if err != nil {
		fmt.Printf("Error generating maze: %v\n", err)
		return
	}
	fmt.Printf("Maze generated with dimensions %dx%d\n", width, height)
}
