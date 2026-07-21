package draw

import (
	"image/color"

	ebit "github.com/hajimehoshi/ebiten/v2"
)

func NewMatrix(width, height int) [][]int {
	matrix := make([][]int, height)
	for i := range matrix {
		matrix[i] = make([]int, width)
	}
	return matrix
}

// DrawScaled renders sprite into cachedImg via buf, then draws it onto dst at (x, y)
// scaled uniformly by scale. cachedImg and buf must match the sprite's natural dimensions.
func DrawScaled(dst *ebit.Image, cachedImg *ebit.Image, buf []byte, sprite *ColorMatrix, x, y float64, scale float64) {
	FillPixelBuffer(buf, sprite.Render())
	cachedImg.WritePixels(buf)
	op := &ebit.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	dst.DrawImage(cachedImg, op)
}

// FillPixelBuffer writes rendered pixel data into a pre-allocated RGBA byte buffer.
// buf must be len(pixels[0]) * len(pixels) * 4 bytes.
func FillPixelBuffer(buf []byte, pixels [][]color.RGBA) {
	if len(pixels) == 0 || len(pixels[0]) == 0 {
		return
	}
	w := len(pixels[0])
	for y, row := range pixels {
		for x, c := range row {
			i := (y*w + x) * 4
			buf[i] = c.R
			buf[i+1] = c.G
			buf[i+2] = c.B
			buf[i+3] = c.A
		}
	}
}
