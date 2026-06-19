package draw

import (
	"fmt"
	"image/color"
)

var (
	ErrInvalidMatrix = fmt.Errorf("invalid matrix: must be non-empty and rectangular")
	ErrKeyCollision = fmt.Errorf("color code key collision")
	ErrIncompatibleDimensions = fmt.Errorf("incompatible dimensions for operation")
)



type AnimationSequence struct {
	Frames		  []color.RGBA 
	FrameDuration int          
	CurrentFrame   int        
	CurrentTick  int
}

func NewAnimationSequence(frames []color.RGBA, frameDuration int) *AnimationSequence {
	if frameDuration <= 0 {
		frameDuration = 1 // Default to 1 if invalid
	}

	return &AnimationSequence{
		Frames:     frames,
		FrameDuration: frameDuration,
		CurrentFrame:   0,
		CurrentTick:   0,
	}
}

func (a *AnimationSequence) GetColor() color.RGBA {
	if len(a.Frames) == 0 {
		return color.RGBA{0, 0, 0, 0} // Return transparent if no frames
	}

	return a.Frames[a.CurrentFrame]
}

func (a *AnimationSequence) Advance() {
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
	Matrix [][]int
	ColorCodes map[int]color.RGBA
	animationSequences map[int]*AnimationSequence
}

func NewColorMatrix(matrix [][]int, colorCodes map[int]color.RGBA, animationSequences map[int]*AnimationSequence) (*ColorMatrix, error) {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return nil, ErrInvalidMatrix
	}

	for _, row := range matrix {
		if len(row) != len(matrix[0]) {
			return nil, ErrInvalidMatrix
		}
	}

	for colorCode := range colorCodes {
		if _, exists := animationSequences[colorCode]; exists {
			return nil, ErrKeyCollision
		}
	}

	for colorCode := range animationSequences {
		if _, exists := colorCodes[colorCode]; exists {
			return nil, ErrKeyCollision
		}
	}

	return &ColorMatrix{
		Matrix: matrix,
		ColorCodes: colorCodes,
		animationSequences: animationSequences,
	}, nil
}



func (cm *ColorMatrix) Render() [][]color.RGBA {
	height := len(cm.Matrix)
	width := len(cm.Matrix[0])
	rendered := make([][]color.RGBA, height)
	for i := range rendered {
		rendered[i] = make([]color.RGBA, width)
	}

	for row := range height {	
		for col := range width {
			colorCode := cm.Matrix[row][col]
			if animSeq, exists := cm.animationSequences[colorCode]; exists {
				rendered[row][col] = animSeq.GetColor()
			} else if colorCode, exists := cm.ColorCodes[colorCode]; exists {
				rendered[row][col] = colorCode
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

func (cm *ColorMatrix) AppendRight(other *ColorMatrix) error {
	if len(cm.Matrix) != len(other.Matrix) {
		return ErrIncompatibleDimensions
	}

	// We assume the other matrix has the same color codes and animation sequences, so we can just append the rows
	for i := range cm.Matrix {
		cm.Matrix[i] = append(cm.Matrix[i], other.Matrix[i]...)
	}

	return nil
}

func (cm *ColorMatrix) AppendBelow(other *ColorMatrix) error {
	if len(cm.Matrix[0]) != len(other.Matrix[0]) {
		return ErrIncompatibleDimensions
	}

	// We assume the other matrix has the same color codes and animation sequences, so we can just append the rows
	cm.Matrix = append(cm.Matrix, other.Matrix...)

	return nil
}

func (cm *ColorMatrix) Compose(other *ColorMatrix, offsetX, offsetY int) error {
	// First reindex the colors and animations of the other matrix to avoid collisions
	reindex := map[int]int{}
	maxCode := 0
	for code := range cm.ColorCodes {
		if code > maxCode {
			maxCode = code
		}
	}
	for code := range cm.animationSequences {
		if code > maxCode {
			maxCode = code
		}
	}
	
	// Start by checking if the other matrix encodes the same color codes
	colorToCode := map[color.RGBA]int{}
	for code, colorValue := range cm.ColorCodes {
		colorToCode[colorValue] = code
	}
	for code, colorValue := range other.ColorCodes {
		if existingCode, exists := colorToCode[colorValue]; exists {
			reindex[code] = existingCode
		} else {
			maxCode++
			reindex[code] = maxCode
			cm.ColorCodes[maxCode] = colorValue
		}
	}

	// We assume all animation sequences are unique, so we reindex them as well
	for code, animSeq := range other.animationSequences {
		maxCode++
		reindex[code] = maxCode
		cm.animationSequences[maxCode] = animSeq
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
				if reindexedCode, exists := reindex[otherCode]; exists {
					cm.Matrix[newRow][newCol] = reindexedCode
				}
			}
		}
	}

	return nil
}