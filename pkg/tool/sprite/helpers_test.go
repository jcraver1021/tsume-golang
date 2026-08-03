package sprite_test

import (
	"image/color"
	"testing"

	. "tsumegolang/pkg/tool/sprite"
)

// makeSolidSprite returns a rows×cols Sprite where every pixel is color c.
func makeSolidSprite(rows, cols int, c color.RGBA) *Sprite {
	p := NewPalette()
	ck, _ := p.Add(c)
	matrix := make([][]ColorKey, rows)
	for i := range matrix {
		matrix[i] = make([]ColorKey, cols)
		for j := range matrix[i] {
			matrix[i][j] = ck
		}
	}
	s, err := NewSprite(matrix, p, map[ColorKey]*AnimationSequence{})
	if err != nil {
		panic(err)
	}
	return s
}

// animationSprite builds a 1×1 Sprite whose single pixel is driven by an
// AnimationSequence over the given colors at the given frameDuration.
func animationSprite(t *testing.T, colors []color.RGBA, frameDuration int) *Sprite {
	t.Helper()
	palette := NewPalette()
	frames := make([]ColorKey, len(colors))
	for i, c := range colors {
		ck, _ := palette.Add(c)
		frames[i] = ck
	}
	keyA, _ := palette.Reserve("A")
	seq, err := NewAnimationSequence(palette, frames, frameDuration)
	if err != nil {
		t.Fatalf("NewAnimationSequence: %v", err)
	}
	sprite, err := NewSprite(
		[][]ColorKey{{keyA}},
		palette,
		map[ColorKey]*AnimationSequence{keyA: seq},
	)
	if err != nil {
		t.Fatalf("NewSprite: %v", err)
	}
	return sprite
}
