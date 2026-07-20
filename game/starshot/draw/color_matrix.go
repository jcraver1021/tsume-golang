package draw

import (
	"fmt"
	"image/color"
)

var (
	ErrInvalidColorKey        = fmt.Errorf("invalid color key: must be a single character")
	ErrInvalidMatrix          = fmt.Errorf("invalid matrix: must be non-empty and rectangular")
	ErrKeyCollision           = fmt.Errorf("color code key collision")
	ErrIncompatibleDimensions = fmt.Errorf("incompatible dimensions for operation")
)

type ColorKey string

func (ck ColorKey) valid() bool {
	return len(ck) == 1 // ColorKey should be a single character
}

func fromString(s string) (ColorKey, error) {
	ck := ColorKey(s)
	if !ck.valid() {
		return "", ErrInvalidColorKey
	}

	return ck, nil
}

type ColorMap map[ColorKey]color.RGBA

type AnimationSequence struct {
	ColorMap      *ColorMap
	Frames        []ColorKey
	FrameDuration int
	CurrentFrame  int
	CurrentTick   int
}

func NewAnimationSequence(colorMap *ColorMap, frames []ColorKey, frameDuration int) *AnimationSequence {
	if frameDuration <= 0 {
		frameDuration = 1 // Default to 1 if invalid
	}

	return &AnimationSequence{
		ColorMap:      colorMap,
		Frames:        frames,
		FrameDuration: frameDuration,
		CurrentFrame:  0,
		CurrentTick:   0,
	}
}

func (a *AnimationSequence) GetColor() color.RGBA {
	if len(a.Frames) == 0 {
		return color.RGBA{0, 0, 0, 0} // Return transparent if no frames
	}

	return (*a.ColorMap)[a.Frames[a.CurrentFrame]]
}

func (a *AnimationSequence) Advance() {
	// We avoid modulo for performance
	a.CurrentTick++
	if a.CurrentTick >= a.FrameDuration {
		a.CurrentTick = 0
		a.CurrentFrame++
		if a.CurrentFrame >= len(a.Frames) {
			a.CurrentFrame = 0
		}
	}
}

type ColorMatrix struct {
	Matrix             [][]ColorKey
	ColorCodes         *ColorMap
	animationSequences map[ColorKey]*AnimationSequence
}

// BlankColorMatrix returns an all-transparent ColorMatrix of the given dimensions.
// Use it as a composition canvas when you need to control layer order explicitly.
func BlankColorMatrix(width, height int) *ColorMatrix {
	transparent := ColorMap{"0": {0, 0, 0, 0}}
	matrix := make([][]ColorKey, height)
	for r := range matrix {
		matrix[r] = make([]ColorKey, width)
	}
	cm, _ := NewColorMatrix(matrix, &transparent, nil)
	return cm
}

func NewColorMatrix(matrix [][]ColorKey, colorCodes *ColorMap, animationSequences map[ColorKey]*AnimationSequence) (*ColorMatrix, error) {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return nil, ErrInvalidMatrix
	}

	for _, row := range matrix {
		if len(row) != len(matrix[0]) {
			return nil, ErrInvalidMatrix
		}
	}

	for colorCode := range *colorCodes {
		if _, exists := animationSequences[colorCode]; exists {
			return nil, ErrKeyCollision
		}
	}

	for colorCode := range animationSequences {
		if _, exists := (*colorCodes)[colorCode]; exists {
			return nil, ErrKeyCollision
		}
	}

	return &ColorMatrix{
		Matrix:             matrix,
		ColorCodes:         colorCodes,
		animationSequences: animationSequences,
	}, nil
}

func (cm *ColorMatrix) Width() int {
	if len(cm.Matrix) == 0 {
		return 0
	}

	return len(cm.Matrix[0])
}

func (cm *ColorMatrix) Height() int {
	return len(cm.Matrix)
}

func (cm *ColorMatrix) Render() [][]color.RGBA {
	height := cm.Height()
	width := cm.Width()
	rendered := make([][]color.RGBA, height)
	for i := range rendered {
		rendered[i] = make([]color.RGBA, width)
	}

	for row := range height {
		for col := range width {
			colorCode := cm.Matrix[row][col]
			if animSeq, exists := cm.animationSequences[colorCode]; exists {
				rendered[row][col] = animSeq.GetColor()
			} else if colorValue, exists := (*cm.ColorCodes)[colorCode]; exists {
				rendered[row][col] = colorValue
			} else {
				rendered[row][col] = color.RGBA{0, 0, 0, 0}
			}
		}
	}

	for _, animSeq := range cm.animationSequences {
		animSeq.Advance()
	}

	return rendered
}

// Dimensions returns the width and height of the color matrix
func (cm *ColorMatrix) Dimensions() (width, height int) {
	if len(cm.Matrix) == 0 {
		return 0, 0
	}
	return len(cm.Matrix[0]), len(cm.Matrix)
}

func (cm *ColorMatrix) appendRight(other *ColorMatrix) error {
	if len(cm.Matrix) != len(other.Matrix) {
		return ErrIncompatibleDimensions
	}

	// We assume the other matrix has the same color codes and animation sequences, so we can just append the rows
	for i := range cm.Matrix {
		cm.Matrix[i] = append(cm.Matrix[i], other.Matrix[i]...)
	}

	return nil
}

func (cm *ColorMatrix) appendBelow(other *ColorMatrix) error {
	if len(cm.Matrix[0]) != len(other.Matrix[0]) {
		return ErrIncompatibleDimensions
	}

	// We assume the other matrix has the same color codes and animation sequences, so we can just append the rows
	cm.Matrix = append(cm.Matrix, other.Matrix...)

	return nil
}

func (cm *ColorMatrix) Compose(other *ColorMatrix, offsetX, offsetY int) error {
	// Initialize maps if nil
	if cm.animationSequences == nil {
		cm.animationSequences = make(map[ColorKey]*AnimationSequence)
	}

	// First reindex the colors and animations of the other matrix to avoid collisions
	reindex := map[ColorKey]ColorKey{}
	maxCode := 0
	for code := range *cm.ColorCodes {
		codeNum := int(code[0])
		if codeNum > maxCode {
			maxCode = codeNum
		}
	}
	for code := range cm.animationSequences {
		codeNum := int(code[0])
		if codeNum > maxCode {
			maxCode = codeNum
		}
	}

	// Start by checking if the other matrix encodes the same color codes
	colorToCode := map[color.RGBA]ColorKey{}
	for code, colorValue := range *cm.ColorCodes {
		colorToCode[colorValue] = code
	}

	// Handle blending for semi-transparent pixels
	blendedColors := make(map[[2]ColorKey]ColorKey) // [baseCode, overlayCode] -> blendedCode

	for code, colorValue := range *other.ColorCodes {
		if existingCode, exists := colorToCode[colorValue]; exists {
			reindex[code] = existingCode
		} else {
			maxCode++
			newKey := ColorKey(string(rune(maxCode)))
			reindex[code] = newKey
			(*cm.ColorCodes)[newKey] = colorValue
		}
	}

	// We assume all animation sequences are unique, so we reindex them as well
	for code, animSeq := range other.animationSequences {
		maxCode++
		newKey := ColorKey(string(rune(maxCode)))
		reindex[code] = newKey
		cm.animationSequences[newKey] = animSeq
	}

	// Now we can add in the reindexed other matrix into this one at the specified offset
	for row := range other.Matrix {
		for col := range other.Matrix[row] {
			newRow := row + offsetY
			newCol := col + offsetX
			// Note that it is OK to have a negative offset or an offset that goes beyond the bounds of the current matrix;
			// this allows partial overlaps.
			// We just skip any pixels that are out of bounds.
			if newRow >= 0 && newRow < len(cm.Matrix) && newCol >= 0 && newCol < len(cm.Matrix[0]) {
				otherCode := other.Matrix[row][col]

				// Get the overlay color
				var overlayColor color.RGBA
				var hasOverlay bool
				if animSeq, hasAnim := other.animationSequences[otherCode]; hasAnim {
					if len(animSeq.Frames) > 0 {
						overlayColor = (*animSeq.ColorMap)[animSeq.Frames[0]]
						hasOverlay = true
					}
				} else if colorValue, hasColor := (*other.ColorCodes)[otherCode]; hasColor {
					overlayColor = colorValue
					hasOverlay = true
				}

				if !hasOverlay || overlayColor.A == 0 {
					// Fully transparent - don't change base
					continue
				}

				if overlayColor.A == 255 {
					// Fully opaque - simple overwrite
					if reindexedCode, exists := reindex[otherCode]; exists {
						cm.Matrix[newRow][newCol] = reindexedCode
					}
					continue
				}

				// Semi-transparent - need to blend with base
				baseCode := cm.Matrix[newRow][newCol]

				// Check if we already computed this blend
				blendKey := [2]ColorKey{baseCode, otherCode}
				if blendedCode, exists := blendedColors[blendKey]; exists {
					cm.Matrix[newRow][newCol] = blendedCode
					continue
				}

				// Get base color
				var baseColor color.RGBA
				if animSeq, hasAnim := cm.animationSequences[baseCode]; hasAnim {
					if len(animSeq.Frames) > 0 {
						baseColor = (*animSeq.ColorMap)[animSeq.Frames[0]]
					}
				} else if colorValue, hasColor := (*cm.ColorCodes)[baseCode]; hasColor {
					baseColor = colorValue
				}

				// Alpha composite: overlay OVER base
				blended := alphaComposite(overlayColor, baseColor)

				// Create new color code for blended result
				if existingCode, exists := colorToCode[blended]; exists {
					cm.Matrix[newRow][newCol] = existingCode
					blendedColors[blendKey] = existingCode
				} else {
					maxCode++
					newKey := ColorKey(string(rune(maxCode)))
					(*cm.ColorCodes)[newKey] = blended
					colorToCode[blended] = newKey
					cm.Matrix[newRow][newCol] = newKey
					blendedColors[blendKey] = newKey
				}
			}
		}
	}

	return nil
}

// ComposeExpanding overlays another ColorMatrix on top of this one, expanding the base matrix if needed.
// The overlay is centered on the base matrix. If the overlay is larger, the base matrix grows to accommodate it.
// This is useful for death animations where an explosion (larger) needs to overlay a ship (smaller).
func (cm *ColorMatrix) ComposeExpanding(other *ColorMatrix) error {
	baseWidth := cm.Width()
	baseHeight := cm.Height()
	overlayWidth := other.Width()
	overlayHeight := other.Height()

	// Calculate new dimensions (max of both)
	newWidth := max(baseWidth, overlayWidth)
	newHeight := max(baseHeight, overlayHeight)

	// Calculate offsets to center the smaller sprite
	baseOffsetX := (newWidth - baseWidth) / 2
	baseOffsetY := (newHeight - baseHeight) / 2
	overlayOffsetX := (newWidth - overlayWidth) / 2
	overlayOffsetY := (newHeight - overlayHeight) / 2

	// Create new matrix with expanded dimensions
	newMatrix := make([][]ColorKey, newHeight)
	for i := range newMatrix {
		newMatrix[i] = make([]ColorKey, newWidth)
	}

	// Copy base matrix into new matrix at centered position
	for row := range cm.Matrix {
		for col := range cm.Matrix[row] {
			newMatrix[row+baseOffsetY][col+baseOffsetX] = cm.Matrix[row][col]
		}
	}

	// Update cm.Matrix to the expanded matrix
	cm.Matrix = newMatrix

	// Now compose the overlay at its centered position
	return cm.Compose(other, overlayOffsetX, overlayOffsetY)
}

// alphaComposite performs Porter-Duff "source over destination" alpha compositing
// Returns the result of compositing src over dst
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
