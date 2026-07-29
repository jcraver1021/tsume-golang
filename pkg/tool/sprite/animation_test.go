package sprite_test

import (
	"image/color"
	"testing"

	. "tsumegolang/pkg/tool/sprite"
)

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

func TestAnimationSequence(t *testing.T) {
	testCases := []struct {
		name          string
		colors        []color.RGBA
		frameDuration int
		want          []color.RGBA
	}{
		{
			name: "simple sequence",
			colors: []color.RGBA{
				{255, 0, 0, 255},
				{0, 255, 0, 255},
				{0, 0, 255, 255},
			},
			frameDuration: 1,
			want: []color.RGBA{
				{255, 0, 0, 255},
				{0, 255, 0, 255},
				{0, 0, 255, 255},
			},
		},
		{
			name: "sequence with frame duration 2",
			colors: []color.RGBA{
				{255, 0, 0, 255},
				{0, 255, 0, 255},
				{0, 0, 255, 255},
			},
			frameDuration: 2,
			want: []color.RGBA{
				{255, 0, 0, 255}, {255, 0, 0, 255},
				{0, 255, 0, 255}, {0, 255, 0, 255},
				{0, 0, 255, 255}, {0, 0, 255, 255},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sprite := animationSprite(t, tc.colors, tc.frameDuration)

			check := func(label string, want []color.RGBA) {
				for i, wantColor := range want {
					got := sprite.Render()[0][0]
					if got != wantColor {
						t.Errorf("%s step %d: got %v, want %v", label, i, got, wantColor)
					}
					sprite.Advance()
				}
			}

			check("first pass", tc.want)
			check("loop", tc.want) // verify the sequence wraps correctly
		})
	}
}

func TestAnimationSequence_Failure(t *testing.T) {
	_, err := NewAnimationSequence(NewPalette(), []ColorKey{}, 1)
	if err == nil {
		t.Fatalf("Expected error for empty frames, got nil")
	}

	_, err = NewAnimationSequence(NewPalette(), []ColorKey{"r"}, 0)
	if err == nil {
		t.Fatalf("Expected error for invalid frame duration, got nil")
	}
}
