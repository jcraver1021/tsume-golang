package sprite

import (
	"image/color"
)

// alphaComposite performs Porter-Duff "source over destination" alpha compositing and returns the resulting color.
func alphaComposite(src, dst color.RGBA) color.RGBA {
	if src.A == 255 {
		return src // Fully opaque source
	}
	if src.A == 0 {
		return dst // Fully transparent source
	}
	if dst.A == 0 {
		return src // Transparent destination
	}

	// Convert to float for calculations
	srcA := float32(src.A) / 255.0
	dstA := float32(dst.A) / 255.0

	// Result alpha
	outA := srcA + dstA*(1-srcA)

	// Avoid division by zero
	if outA == 0 {
		return color.RGBA{0, 0, 0, 0}
	}

	// Result color channels
	outR := (float32(src.R)*srcA + float32(dst.R)*dstA*(1-srcA)) / outA
	outG := (float32(src.G)*srcA + float32(dst.G)*dstA*(1-srcA)) / outA
	outB := (float32(src.B)*srcA + float32(dst.B)*dstA*(1-srcA)) / outA

	return color.RGBA{
		R: uint8(outR),
		G: uint8(outG),
		B: uint8(outB),
		A: uint8(outA * 255.0),
	}
}
