package draw_test

import (
	"errors"
	"image/color"
	"testing"

	. "tsumegolang/game/starshot/draw"
)

func TestComposeExpanding(t *testing.T) {
	// Create a small base sprite (2×2)
	baseMatrix := [][]ColorKey{
		{"1", "1"},
		{"1", "1"},
	}
	baseColors := ColorMap{
		"1": {255, 0, 0, 255}, // Red
	}
	base, err := NewColorMatrix(baseMatrix, &baseColors, nil)
	if err != nil {
		t.Fatalf("Failed to create base matrix: %v", err)
	}

	// Create a larger overlay sprite (4×4)
	overlayMatrix := [][]ColorKey{
		{"2", "2", "2", "2"},
		{"2", "0", "0", "2"},
		{"2", "0", "0", "2"},
		{"2", "2", "2", "2"},
	}
	overlayColors := ColorMap{
		"2": {0, 0, 255, 255}, // Blue
		"0": {0, 0, 0, 0},     // Transparent
	}
	overlay, err := NewColorMatrix(overlayMatrix, &overlayColors, nil)
	if err != nil {
		t.Fatalf("Failed to create overlay matrix: %v", err)
	}

	// Compose expanding
	err = base.ComposeExpanding(overlay)
	if err != nil {
		t.Fatalf("ComposeExpanding failed: %v", err)
	}

	// Base should now be 4×4
	if base.Width() != 4 || base.Height() != 4 {
		t.Errorf("Expected expanded size 4×4, got %d×%d", base.Width(), base.Height())
	}

	// Center should have the original red sprite (centered in 4×4 = offset 1,1)
	// Edges should have blue from overlay
	rendered := base.Render()

	// Check corners (should be blue from overlay)
	if rendered[0][0] != (color.RGBA{0, 0, 255, 255}) {
		t.Errorf("Top-left corner should be blue, got %v", rendered[0][0])
	}

	// Check center (should be red from base, centered at 1,1 and 2,2)
	if rendered[1][1] != (color.RGBA{255, 0, 0, 255}) {
		t.Errorf("Center should be red, got %v", rendered[1][1])
	}
}

func TestAnimationSequence(t *testing.T) {
	testCases := []struct {
		name          string
		frames        []ColorKey
		colors        ColorMap
		frameDuration int
	}{
		{
			"SingleFrame",
			[]ColorKey{"r"},
			ColorMap{"r": {255, 0, 0, 255}},
			1,
		},
		{
			"TwoFrames",
			[]ColorKey{"r", "g"},
			ColorMap{"r": {255, 0, 0, 255}, "g": {0, 255, 0, 255}},
			2,
		},
		{
			"ThreeFrames",
			[]ColorKey{"r", "g", "b"},
			ColorMap{"r": {255, 0, 0, 255}, "g": {0, 255, 0, 255}, "b": {0, 0, 255, 255}},
			3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			animSeq := NewAnimationSequence(&tc.colors, tc.frames, tc.frameDuration)

			for i := 0; i < len(tc.frames)*tc.frameDuration; i++ {
				wantColor := tc.colors[tc.frames[i/tc.frameDuration]]
				gotColor := animSeq.GetColor()

				if gotColor != wantColor {
					t.Errorf("GetColor() = %v, want %v", gotColor, wantColor)
				}

				// make sure that calling GetColor() multiple times without advancing returns the same color
				gotColorSecondCall := animSeq.GetColor()
				if gotColorSecondCall != wantColor {
					t.Errorf("GetColor() on second call = %v, want %v", gotColorSecondCall, wantColor)
				}

				animSeq.Advance()
			}

			// After completing the full cycle, it should loop back to the first frame
			wantColor := tc.colors[tc.frames[0]]
			gotColor := animSeq.GetColor()
			if gotColor != wantColor {
				t.Errorf("After full cycle, GetColor() = %v, want %v", gotColor, wantColor)
			}
		})
	}
}

func TestColorMatrixCreation(t *testing.T) {
	testCases := []struct {
		name          string
		matrix        [][]ColorKey
		colorCodes    ColorMap
		animationSeqs map[ColorKey]*AnimationSequence
		wantError     error
	}{
		{
			name: "ValidMatrix",
			matrix: [][]ColorKey{
				{"1", "2"},
				{"3", "4"},
			},
			colorCodes: ColorMap{
				"1": {255, 0, 0, 255},
				"2": {0, 255, 0, 255},
				"3": {0, 0, 255, 255},
				"4": {255, 255, 0, 255},
			},
			animationSeqs: map[ColorKey]*AnimationSequence{},
			wantError:     nil,
		},
		{
			name:          "EmptyMatrix",
			matrix:        [][]ColorKey{},
			colorCodes:    ColorMap{},
			animationSeqs: map[ColorKey]*AnimationSequence{},
			wantError:     ErrInvalidMatrix,
		},
		{
			name: "NonRectangularMatrix",
			matrix: [][]ColorKey{
				{"1", "2"},
				{"3"},
			},
			colorCodes:    ColorMap{},
			animationSeqs: map[ColorKey]*AnimationSequence{},
			wantError:     ErrInvalidMatrix,
		},
		{
			name: "KeyCollision",
			matrix: [][]ColorKey{
				{"1", "2"},
				{"3", "4"},
			},
			colorCodes: ColorMap{
				"1": {255, 0, 0, 255},
			},
			animationSeqs: func() map[ColorKey]*AnimationSequence {
				cm := ColorMap{"a": {0, 255, 0, 255}}
				return map[ColorKey]*AnimationSequence{
					"1": NewAnimationSequence(&cm, []ColorKey{"a"}, 1),
				}
			}(),
			wantError: ErrKeyCollision,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cm, err := NewColorMatrix(tc.matrix, &tc.colorCodes, tc.animationSeqs)
			if !errors.Is(err, tc.wantError) {
				t.Errorf("NewColorMatrix() error = %v, want %v", err, tc.wantError)
			}
			if err == nil && cm == nil {
				t.Errorf("NewColorMatrix() returned nil ColorMatrix without error")
			}
		})
	}
}

func TestColorMatrixRender(t *testing.T) {
	testCases := []struct {
		name                string
		matrix              [][]ColorKey
		colorCodes          ColorMap
		animationSeqs       map[ColorKey]*AnimationSequence
		wantRenderedInOrder [][][]color.RGBA
	}{
		{
			name: "SimpleMatrixNoAnimation",
			matrix: [][]ColorKey{
				{"1", "2"},
				{"3", "4"},
			},
			colorCodes: ColorMap{
				"1": {255, 0, 0, 255},
				"2": {0, 255, 0, 255},
				"3": {0, 0, 255, 255},
				"4": {255, 255, 0, 255},
			},
			animationSeqs: map[ColorKey]*AnimationSequence{},
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{255, 0, 0, 255}, {0, 255, 0, 255}},
					{{0, 0, 255, 255}, {255, 255, 0, 255}},
				},
				{
					{{255, 0, 0, 255}, {0, 255, 0, 255}},
					{{0, 0, 255, 255}, {255, 255, 0, 255}},
				},
			},
		},
		{
			name: "SimpleMatrixWithAnimation",
			matrix: [][]ColorKey{
				{"1", "2"},
				{"3", "4"},
			},
			colorCodes: ColorMap{
				"1": {255, 0, 0, 255},
				"3": {0, 0, 255, 255},
				"4": {255, 255, 0, 255},
			},
			animationSeqs: func() map[ColorKey]*AnimationSequence {
				cm := ColorMap{"g": {0, 255, 0, 255}, "G": {0, 128, 0, 255}}
				return map[ColorKey]*AnimationSequence{
					"2": NewAnimationSequence(&cm, []ColorKey{"g", "G"}, 1),
				}
			}(),
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{255, 0, 0, 255}, {0, 255, 0, 255}},
					{{0, 0, 255, 255}, {255, 255, 0, 255}},
				},
				{
					{{255, 0, 0, 255}, {0, 128, 0, 255}},
					{{0, 0, 255, 255}, {255, 255, 0, 255}},
				},
				{
					{{255, 0, 0, 255}, {0, 255, 0, 255}},
					{{0, 0, 255, 255}, {255, 255, 0, 255}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cm, err := NewColorMatrix(tc.matrix, &tc.colorCodes, tc.animationSeqs)
			if err != nil {
				t.Fatalf("NewColorMatrix() error = %v", err)
			}

			for _, expectedFrame := range tc.wantRenderedInOrder {
				rendered := cm.Render() // Render should advance the animation sequences internally
				if !renderingsAreEqual(rendered, expectedFrame) {
					t.Errorf("Render() = %v, want %v", rendered, expectedFrame)
				}
			}
		})
	}
}

func renderingsAreEqual(a, b [][]color.RGBA) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}

		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}

	return true
}
