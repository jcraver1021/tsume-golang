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

	m := image.NewRGBA(image.Rect(0, 0, width*SquareSize, height*SquareSize))
	draw.Draw(m, m.Bounds(), &image.Uniform{color.White}, image.Point{0, 0}, draw.Src)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			current := y*width + x
			rectX := x * SquareSize
			rectY := y * SquareSize

			if x == 0 {
				draw.Draw(m, image.Rect(rectX, rectY, rectX+1, rectY+SquareSize), &image.Uniform{color.Black}, image.Point{}, draw.Src)
			}
			if y == 0 && x != 0 {
				draw.Draw(m, image.Rect(rectX, rectY, rectX+SquareSize, rectY+1), &image.Uniform{color.Black}, image.Point{}, draw.Src)
			}
			if x == width-1 {
				draw.Draw(m, image.Rect(rectX+SquareSize-1, rectY, rectX+SquareSize, rectY+SquareSize), &image.Uniform{color.Black}, image.Point{}, draw.Src)
			}
			if y == height-1 && x != width-1 {
				draw.Draw(m, image.Rect(rectX, rectY+SquareSize-1, rectX+SquareSize, rectY+SquareSize), &image.Uniform{color.Black}, image.Point{}, draw.Src)
			}
			if x < width-1 { // East neighbor
				east := current + 1
				if _, connected := g.GetEdge(current, east); !connected {
					draw.Draw(m, image.Rect(rectX+SquareSize-1, rectY, rectX+SquareSize, rectY+SquareSize), &image.Uniform{color.Black}, image.Point{}, draw.Src)
				}
			}
			if y < height-1 { // South neighbor
				south := current + width
				if _, connected := g.GetEdge(current, south); !connected {
					draw.Draw(m, image.Rect(rectX, rectY+SquareSize-1, rectX+SquareSize, rectY+SquareSize), &image.Uniform{color.Black}, image.Point{}, draw.Src)
				}
			}
		}
	}

	return m, nil
}
