package draw_test

import (
	"errors"
	"image/color"
	"testing"

	. "tsumegolang/game/starshot/draw"
)

func TestAnimationSequence(t *testing.T) {
	testCases := []struct {
		name          string
		sequence      []color.RGBA
		frameDuration int
	}{
		{"SingleFrame", []color.RGBA{{255, 0, 0, 255}}, 1},
		{"TwoFrames", []color.RGBA{{255, 0, 0, 255}, {0, 255, 0, 255}}, 2},
		{"ThreeFrames", []color.RGBA{{255, 0, 0, 255}, {0, 255, 0, 255}, {0, 0, 255, 255}}, 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			animSeq := NewAnimationSequence(tc.sequence, tc.frameDuration)

			for i := 0; i < len(tc.sequence)*tc.frameDuration; i++ {
				wantColor := tc.sequence[i/tc.frameDuration]
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
			wantColor := tc.sequence[0]
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
		matrix        [][]int
		colorCodes    map[int]color.RGBA
		animationSeqs map[int]*AnimationSequence
		wantError     error
	}{
		{
			name: "ValidMatrix",
			matrix: [][]int{
				{1, 2},
				{3, 4},
			},
			colorCodes: map[int]color.RGBA{
				1: {255, 0, 0, 255},
				2: {0, 255, 0, 255},
				3: {0, 0, 255, 255},
				4: {255, 255, 0, 255},
			},
			animationSeqs: map[int]*AnimationSequence{},
			wantError:     nil,
		},
		{
			name:          "EmptyMatrix",
			matrix:        [][]int{},
			colorCodes:    map[int]color.RGBA{},
			animationSeqs: map[int]*AnimationSequence{},
			wantError:     ErrInvalidMatrix,
		},
		{
			name: "NonRectangularMatrix",
			matrix: [][]int{
				{1, 2},
				{3},
			},
			colorCodes:    map[int]color.RGBA{},
			animationSeqs: map[int]*AnimationSequence{},
			wantError:     ErrInvalidMatrix,
		},
		{
			name: "KeyCollision",
			matrix: [][]int{
				{1, 2},
				{3, 4},
			},
			colorCodes: map[int]color.RGBA{
				1: {255, 0, 0, 255},
			},
			animationSeqs: map[int]*AnimationSequence{
				1: NewAnimationSequence([]color.RGBA{{0, 255, 0, 255}}, 1),
			},
			wantError: ErrKeyCollision,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cm, err := NewColorMatrix(tc.matrix, tc.colorCodes, tc.animationSeqs)
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
		matrix              [][]int
		colorCodes          map[int]color.RGBA
		animationSeqs       map[int]*AnimationSequence
		wantRenderedInOrder [][][]color.RGBA
	}{
		{
			name: "SimpleMatrixNoAnimation",
			matrix: [][]int{
				{1, 2},
				{3, 4},
			},
			colorCodes: map[int]color.RGBA{
				1: {255, 0, 0, 255},
				2: {0, 255, 0, 255},
				3: {0, 0, 255, 255},
				4: {255, 255, 0, 255},
			},
			animationSeqs: map[int]*AnimationSequence{},
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
			matrix: [][]int{
				{1, 2},
				{3, 4},
			},
			colorCodes: map[int]color.RGBA{
				1: {255, 0, 0, 255},
				3: {0, 0, 255, 255},
				4: {255, 255, 0, 255},
			},
			animationSeqs: map[int]*AnimationSequence{
				2: NewAnimationSequence([]color.RGBA{{0, 255, 0, 255}, {0, 128, 0, 255}}, 1),
			},
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
			cm, err := NewColorMatrix(tc.matrix, tc.colorCodes, tc.animationSeqs)
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
