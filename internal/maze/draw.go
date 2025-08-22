package maze

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
)

const (
	SquareSize = 20 // Size of each square in the rectangle drawing
)

func DrawRectangleMaze(r *Rectangle) (image.Image, error) {
	g := r.Graph
	width := r.Width
	height := r.Height
	if g.GetSize() != width*height {
		return nil, fmt.Errorf("graph size %d does not match specified dimensions %dx%d", g.GetSize(), width, height)
	}

	// Create a white rectangle.
	m := image.NewRGBA(image.Rect(0, 0, width*SquareSize, height*SquareSize))
	draw.Draw(m, m.Bounds(), &image.Uniform{color.White}, image.Point{0, 0}, draw.Src)

	// For each cell, draw borders as lines.
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			current := y*width + x
			rectX := x * SquareSize
			rectY := y * SquareSize

			if x == 0 {
				// Draw left border
				draw.Draw(m, image.Rect(rectX, rectY, rectX+1, rectY+SquareSize), &image.Uniform{color.Black}, image.Point{}, draw.Src)
			}
			if y == 0 && x != 0 {
				// Draw top border except for the first cell
				draw.Draw(m, image.Rect(rectX, rectY, rectX+SquareSize, rectY+1), &image.Uniform{color.Black}, image.Point{}, draw.Src)
			}
			if x == width-1 {
				// Draw right border
				draw.Draw(m, image.Rect(rectX+SquareSize-1, rectY, rectX+SquareSize, rectY+SquareSize), &image.Uniform{color.Black}, image.Point{}, draw.Src)
			}
			if y == height-1 && x != width-1 {
				// Draw bottom border except for the last cell
				draw.Draw(m, image.Rect(rectX, rectY+SquareSize-1, rectX+SquareSize, rectY+SquareSize), &image.Uniform{color.Black}, image.Point{}, draw.Src)
			}
			if x < width-1 { // East neighbor
				east := current + 1
				if _, connected := g.GetEdge(current, east); !connected {
					// Draw vertical line to the east neighbor
					draw.Draw(m, image.Rect(rectX+SquareSize-1, rectY, rectX+SquareSize, rectY+SquareSize), &image.Uniform{color.Black}, image.Point{}, draw.Src)
				}
			}
			if y < height-1 { // South neighbor
				south := current + width
				if _, connected := g.GetEdge(current, south); !connected {
					// Draw horizontal line to the south neighbor
					draw.Draw(m, image.Rect(rectX, rectY+SquareSize-1, rectX+SquareSize, rectY+SquareSize), &image.Uniform{color.Black}, image.Point{}, draw.Src)
				}
			}
		}
	}

	return m, nil
}
