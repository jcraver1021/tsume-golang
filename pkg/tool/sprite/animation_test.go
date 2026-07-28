package sprite_test

import (
	"image/color"
	"testing"

	. "tsumegolang/pkg/tool/sprite"
)

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
		palette := NewPalette()
		frames := make([]ColorKey, len(tc.colors))
		for i, c := range tc.colors {
			ck, _ := palette.Add(c)
			frames[i] = ck
		}

		seq, err := NewAnimationSequence(palette, frames, tc.frameDuration)
		if err != nil {
			t.Fatalf("NewAnimationSequence() error = %v", err)
		}

		for i, got := range tc.want {
			want := seq.GetColor()
			if want != got {
				t.Errorf("Frame %d: got %v, want %v", i, got, want)
			}
			seq.Advance()
		}

		// Repeat the sequence to verify it loops correctly
		for i, got := range tc.want {
			want := seq.GetColor()
			if want != got {
				t.Errorf("Loop Frame %d: got %v, want %v", i, got, want)
			}
			seq.Advance()
		}
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
