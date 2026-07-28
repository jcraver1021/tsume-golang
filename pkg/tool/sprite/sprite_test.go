package sprite_test

import (
	"image/color"
	"testing"

	. "tsumegolang/pkg/tool/sprite"
)

func TestSpriteCreation(t *testing.T) {
	testCases := []struct {
		name               string
		matrix             [][]ColorKey
		palette            *Palette
		animationSequences map[ColorKey]*AnimationSequence
		wantErr            bool
	}{
		{
			name: "ValidSprite",
			matrix: [][]ColorKey{
				{"1", "2"},
				{"3", "4"},
			},
			palette: func() *Palette {
				p := NewPalette()
				p.Add(color.RGBA{R: 255, G: 0, B: 0, A: 255})
				p.Add(color.RGBA{R: 0, G: 255, B: 0, A: 255})
				p.Add(color.RGBA{R: 0, G: 0, B: 255, A: 255})
				p.Add(color.RGBA{R: 255, G: 255, B: 0, A: 255})
				return p
			}(),
			animationSequences: map[ColorKey]*AnimationSequence{},
			wantErr:            false,
		},
		{
			name:               "InvalidSpriteEmptyMatrix",
			matrix:             [][]ColorKey{},
			palette:            NewPalette(),
			animationSequences: map[ColorKey]*AnimationSequence{},
			wantErr:            true,
		},
		{
			name: "InvalidSpriteNonRectangularMatrix",
			matrix: [][]ColorKey{
				{"1", "2"},
				{"3"},
			},
			palette:            NewPalette(),
			animationSequences: map[ColorKey]*AnimationSequence{},
			wantErr:            true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewSprite(tc.matrix, tc.palette, tc.animationSequences)
			if (err != nil) != tc.wantErr {
				t.Errorf("expected error: %v, got: %v", tc.wantErr, err)
			}
		})
	}
}

func TestBlankSprite(t *testing.T) {
	rows, cols := 2, 3
	sprite, err := BlankSprite(rows, cols)
	if err != nil {
		t.Fatalf("BlankSprite() error = %v", err)
	}
	if sprite.Width() != cols {
		t.Errorf("expected width: %d, got: %d", cols, sprite.Width())
	}
	if sprite.Height() != rows {
		t.Errorf("expected height: %d, got: %d", rows, sprite.Height())
	}
}

func TestSpriteClone(t *testing.T) {
	t.Run("CloneHasSamePixels", func(t *testing.T) {
		original := makeSolidSprite(2, 2, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		clone := original.Clone()
		origFrame := original.Render()
		cloneFrame := clone.Render()
		for r := range origFrame {
			for c := range origFrame[r] {
				if cloneFrame[r][c] != origFrame[r][c] {
					t.Errorf("pixel (%d,%d): original %v, clone %v", r, c, origFrame[r][c], cloneFrame[r][c])
				}
			}
		}
	})

	t.Run("CloneAnimationsAdvanceIndependently", func(t *testing.T) {
		palette := NewPalette()
		key1, _ := palette.Add(color.RGBA{R: 255, G: 0, B: 0, A: 255}) // red
		key2, _ := palette.Add(color.RGBA{R: 0, G: 255, B: 0, A: 255}) // green
		keyA, _ := palette.Reserve("A")
		seq, _ := NewAnimationSequence(palette, []ColorKey{key1, key2}, 1)
		matrix := [][]ColorKey{{keyA}}
		original, _ := NewSprite(matrix, palette, map[ColorKey]*AnimationSequence{keyA: seq})

		clone := original.Clone()

		// Advance original by one frame; clone must be unaffected.
		original.Render() // renders red (frame 0) — pure, no side effect
		original.Advance()

		// Clone should still be at frame 0 (red) since it wasn't advanced.
		cloneColor := clone.Render()[0][0]
		want := color.RGBA{R: 255, G: 0, B: 0, A: 255}
		if cloneColor != want {
			t.Errorf("clone was affected by original's advance: got %v, want %v", cloneColor, want)
		}
	})

	t.Run("ClonePreservesAnimationState", func(t *testing.T) {
		palette := NewPalette()
		key1, _ := palette.Add(color.RGBA{R: 255, G: 0, B: 0, A: 255}) // red
		key2, _ := palette.Add(color.RGBA{R: 0, G: 255, B: 0, A: 255}) // green
		keyA, _ := palette.Reserve("A")
		seq, _ := NewAnimationSequence(palette, []ColorKey{key1, key2}, 1)
		matrix := [][]ColorKey{{keyA}}
		sprite, _ := NewSprite(matrix, palette, map[ColorKey]*AnimationSequence{keyA: seq})

		sprite.Render()  // renders red (frame 0) — pure, no side effect
		sprite.Advance() // step to frame 1 (green)

		clone := sprite.Clone() // clone starts at frame 1

		// Both original and clone should now render green (frame 1).
		origColor := sprite.Render()[0][0]
		cloneColor := clone.Render()[0][0]
		if origColor != cloneColor {
			t.Errorf("frame mismatch after clone: original %v, clone %v", origColor, cloneColor)
		}
		want := color.RGBA{R: 0, G: 255, B: 0, A: 255}
		if origColor != want {
			t.Errorf("expected green at frame 1, got %v", origColor)
		}
	})
}

func TestSpriteRender(t *testing.T) {
	testCases := []struct {
		name                string
		spriteFn            func() *Sprite
		wantRenderedInOrder [][][]color.RGBA
	}{
		{
			name: "BlankSpriteRender",
			spriteFn: func() *Sprite {
				sprite, _ := BlankSprite(2, 3)
				return sprite
			},
			// multiple frames to check stability
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
					{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				},
				{
					{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
					{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}},
				},
			},
		},
		{
			name: "SimpleSpriteNoAnimation",
			spriteFn: func() *Sprite {
				palette := NewPalette()
				key1, _ := palette.Add(color.RGBA{R: 255, G: 0, B: 0, A: 255})
				key2, _ := palette.Add(color.RGBA{R: 0, G: 255, B: 0, A: 255})
				key3, _ := palette.Add(color.RGBA{R: 0, G: 0, B: 255, A: 255})
				key4, _ := palette.Add(color.RGBA{R: 255, G: 255, B: 0, A: 255})
				matrix := [][]ColorKey{
					{key1, key2},
					{key3, key4},
				}
				sprite, _ := NewSprite(matrix, palette, map[ColorKey]*AnimationSequence{})
				return sprite
			},
			// multiple frames to check stability
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
			name: "SimpleSpriteWithAnimation",
			spriteFn: func() *Sprite {
				palette := NewPalette()
				key1, _ := palette.Add(color.RGBA{R: 255, G: 0, B: 0, A: 255})
				key2, _ := palette.Add(color.RGBA{R: 0, G: 255, B: 0, A: 255})
				key3, _ := palette.Add(color.RGBA{R: 0, G: 0, B: 255, A: 255})
				key4, _ := palette.Add(color.RGBA{R: 255, G: 255, B: 0, A: 255})
				keyA, _ := palette.Reserve("A")
				matrix := [][]ColorKey{
					{key1, key2},
					{key3, keyA},
				}
				animationSequences := map[ColorKey]*AnimationSequence{}
				animationSequences[keyA], _ = NewAnimationSequence(palette, []ColorKey{key3, key4}, 2)
				sprite, _ := NewSprite(matrix, palette, animationSequences)
				return sprite
			},
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{255, 0, 0, 255}, {0, 255, 0, 255}},
					{{0, 0, 255, 255}, {0, 0, 255, 255}},
				},
				{
					{{255, 0, 0, 255}, {0, 255, 0, 255}},
					{{0, 0, 255, 255}, {0, 0, 255, 255}},
				},
				{
					{{255, 0, 0, 255}, {0, 255, 0, 255}},
					{{0, 0, 255, 255}, {255, 255, 0, 255}},
				},
				{
					{{255, 0, 0, 255}, {0, 255, 0, 255}},
					{{0, 0, 255, 255}, {255, 255, 0, 255}},
				},
				// duplicated to check animation looping
				{
					{{255, 0, 0, 255}, {0, 255, 0, 255}},
					{{0, 0, 255, 255}, {0, 0, 255, 255}},
				},
				{
					{{255, 0, 0, 255}, {0, 255, 0, 255}},
					{{0, 0, 255, 255}, {0, 0, 255, 255}},
				},
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sprite := tc.spriteFn()
			for i, wantFrame := range tc.wantRenderedInOrder {
				gotFrame := sprite.Render()
				for r := range wantFrame {
					for c := range wantFrame[r] {
						if gotFrame[r][c] != wantFrame[r][c] {
							t.Errorf("frame %d, pixel (%d,%d): expected %v, got %v", i, r, c, wantFrame[r][c], gotFrame[r][c])
						}
					}
				}
				sprite.Advance()
			}
		})
	}
}

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

func TestSpriteComposition(t *testing.T) {
	type spritePair struct {
		dst *Sprite
		src *Sprite
	}
	testCases := []struct {
		name                string
		spritesFn           func() spritePair
		rowOffset           int
		colOffset           int
		wantRenderedInOrder [][][]color.RGBA
	}{
		{
			name: "OpaqueSourceFullOverlap",
			spritesFn: func() spritePair {
				return spritePair{
					dst: makeSolidSprite(2, 2, color.RGBA{R: 255, G: 0, B: 0, A: 255}),
					src: makeSolidSprite(2, 2, color.RGBA{R: 0, G: 255, B: 0, A: 255}),
				}
			},
			rowOffset: 0,
			colOffset: 0,
			// Opaque green src completely covers the red dst.
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{0, 255, 0, 255}, {0, 255, 0, 255}},
					{{0, 255, 0, 255}, {0, 255, 0, 255}},
				},
			},
		},
		{
			name: "OpaqueSourcePartialOverlap",
			spritesFn: func() spritePair {
				return spritePair{
					dst: makeSolidSprite(2, 2, color.RGBA{R: 255, G: 0, B: 0, A: 255}),
					src: makeSolidSprite(1, 1, color.RGBA{R: 0, G: 255, B: 0, A: 255}),
				}
			},
			rowOffset: 0,
			colOffset: 0,
			// Only dst[0][0] is within the 1×1 src; the rest remain red.
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{0, 255, 0, 255}, {255, 0, 0, 255}},
					{{255, 0, 0, 255}, {255, 0, 0, 255}},
				},
			},
		},
		{
			name: "TransparentSourcePreservesDst",
			spritesFn: func() spritePair {
				return spritePair{
					dst: makeSolidSprite(2, 2, color.RGBA{R: 255, G: 0, B: 0, A: 255}),
					src: makeSolidSprite(2, 2, color.RGBA{R: 0, G: 0, B: 0, A: 0}),
				}
			},
			rowOffset: 0,
			colOffset: 0,
			// Fully transparent src leaves every dst pixel untouched.
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{255, 0, 0, 255}, {255, 0, 0, 255}},
					{{255, 0, 0, 255}, {255, 0, 0, 255}},
				},
			},
		},
		{
			name: "SemiTransparentBlend",
			spritesFn: func() spritePair {
				return spritePair{
					dst: makeSolidSprite(1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255}),
					src: makeSolidSprite(1, 1, color.RGBA{R: 0, G: 0, B: 255, A: 128}),
				}
			},
			rowOffset: 0,
			colOffset: 0,
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{126, 0, 128, 255}},
				},
			},
		},
		{
			name: "NegativeOffsetPlacesSourceAtCenter",
			spritesFn: func() spritePair {
				return spritePair{
					dst: makeSolidSprite(3, 3, color.RGBA{R: 255, G: 0, B: 0, A: 255}),
					src: makeSolidSprite(1, 1, color.RGBA{R: 0, G: 255, B: 0, A: 255}),
				}
			},
			rowOffset: -1,
			colOffset: -1,
			// dst[i][j] maps to src[i+offset][j+offset]; offset=-1 places src[0][0] at dst[1][1].
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{255, 0, 0, 255}, {255, 0, 0, 255}, {255, 0, 0, 255}},
					{{255, 0, 0, 255}, {0, 255, 0, 255}, {255, 0, 0, 255}},
					{{255, 0, 0, 255}, {255, 0, 0, 255}, {255, 0, 0, 255}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pair := tc.spritesFn()
			dst := pair.dst
			src := pair.src

			dstBefore := dst.Render()

			result, err := dst.Compose(src, tc.rowOffset, tc.colOffset)
			if err != nil {
				t.Fatalf("Compose failed: %v", err)
			}

			// dst must be pixel-identical before and after the call.
			dstAfter := dst.Render()
			for r := range dstBefore {
				for c := range dstBefore[r] {
					if dstBefore[r][c] != dstAfter[r][c] {
						t.Errorf("dst mutated at (%d,%d): before %v, after %v", r, c, dstBefore[r][c], dstAfter[r][c])
					}
				}
			}

			for i, wantFrame := range tc.wantRenderedInOrder {
				gotFrame := result.Render()
				for r := range wantFrame {
					for c := range wantFrame[r] {
						if gotFrame[r][c] != wantFrame[r][c] {
							t.Errorf("frame %d, pixel (%d,%d): expected %v, got %v", i, r, c, wantFrame[r][c], gotFrame[r][c])
						}
					}
				}
			}
		})
	}
}

func TestSpriteCompositionWithExpansion(t *testing.T) {
	type spritePair struct {
		dst *Sprite
		src *Sprite
	}
	testCases := []struct {
		name                string
		spritesFn           func() spritePair
		wantWidth           int
		wantHeight          int
		wantRenderedInOrder [][][]color.RGBA
	}{
		{
			name: "SameSizeOpaqueOverlay",
			spritesFn: func() spritePair {
				return spritePair{
					dst: makeSolidSprite(2, 2, color.RGBA{R: 255, G: 0, B: 0, A: 255}),
					src: makeSolidSprite(2, 2, color.RGBA{R: 0, G: 255, B: 0, A: 255}),
				}
			},
			wantWidth:  2,
			wantHeight: 2,
			// Equal sizes → offsets both 0; full overlap; green wins everywhere.
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{0, 255, 0, 255}, {0, 255, 0, 255}},
					{{0, 255, 0, 255}, {0, 255, 0, 255}},
				},
			},
		},
		{
			name: "SameSizeTransparentSource",
			spritesFn: func() spritePair {
				return spritePair{
					dst: makeSolidSprite(2, 2, color.RGBA{R: 255, G: 0, B: 0, A: 255}),
					src: makeSolidSprite(2, 2, color.RGBA{R: 0, G: 0, B: 0, A: 0}),
				}
			},
			wantWidth:  2,
			wantHeight: 2,
			// Transparent src leaves dst untouched; no resize needed.
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{255, 0, 0, 255}, {255, 0, 0, 255}},
					{{255, 0, 0, 255}, {255, 0, 0, 255}},
				},
			},
		},
		{
			name: "SameSizeSemiTransparentBlend",
			spritesFn: func() spritePair {
				return spritePair{
					dst: makeSolidSprite(2, 2, color.RGBA{R: 255, G: 0, B: 0, A: 255}),
					src: makeSolidSprite(2, 2, color.RGBA{R: 0, G: 0, B: 255, A: 128}),
				}
			},
			wantWidth:  2,
			wantHeight: 2,
			// Same per-pixel blend as SemiTransparentBlend in TestSpriteComposition.
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{126, 0, 128, 255}, {126, 0, 128, 255}},
					{{126, 0, 128, 255}, {126, 0, 128, 255}},
				},
			},
		},
		{
			name: "SrcSmallerNoExpansionTopLeftPlacement",
			spritesFn: func() spritePair {
				return spritePair{
					dst: makeSolidSprite(2, 2, color.RGBA{R: 255, G: 0, B: 0, A: 255}),
					src: makeSolidSprite(1, 1, color.RGBA{R: 0, G: 255, B: 0, A: 255}),
				}
			},
			wantWidth:  2,
			wantHeight: 2,
			// overlayOffset = (2-1)/2 = 0; src lands at top-left (0,0).
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{0, 255, 0, 255}, {255, 0, 0, 255}},
					{{255, 0, 0, 255}, {255, 0, 0, 255}},
				},
			},
		},
		{
			name: "SrcSmallerCenteredInDst",
			spritesFn: func() spritePair {
				return spritePair{
					dst: makeSolidSprite(3, 3, color.RGBA{R: 255, G: 0, B: 0, A: 255}),
					src: makeSolidSprite(1, 1, color.RGBA{R: 0, G: 255, B: 0, A: 255}),
				}
			},
			wantWidth:  3,
			wantHeight: 3,
			// overlayOffset = (3-1)/2 = 1; negated → Compose(-1,-1) places src[0][0] at dst[1][1].
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{255, 0, 0, 255}, {255, 0, 0, 255}, {255, 0, 0, 255}},
					{{255, 0, 0, 255}, {0, 255, 0, 255}, {255, 0, 0, 255}},
					{{255, 0, 0, 255}, {255, 0, 0, 255}, {255, 0, 0, 255}},
				},
			},
		},
		{
			name: "SrcLargerExpandsDst",
			spritesFn: func() spritePair {
				return spritePair{
					dst: makeSolidSprite(1, 1, color.RGBA{R: 255, G: 0, B: 0, A: 255}),
					src: makeSolidSprite(3, 3, color.RGBA{R: 0, G: 255, B: 0, A: 255}),
				}
			},
			wantWidth:  3,
			wantHeight: 3,
			// dst expands to 3×3; red lands at center (1,1); border cells are
			// transparent and yield to the opaque green src everywhere.
			wantRenderedInOrder: [][][]color.RGBA{
				{
					{{0, 255, 0, 255}, {0, 255, 0, 255}, {0, 255, 0, 255}},
					{{0, 255, 0, 255}, {0, 255, 0, 255}, {0, 255, 0, 255}},
					{{0, 255, 0, 255}, {0, 255, 0, 255}, {0, 255, 0, 255}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pair := tc.spritesFn()
			dst := pair.dst
			src := pair.src

			origDstWidth := dst.Width()
			origDstHeight := dst.Height()
			dstBefore := dst.Render()

			result, err := dst.ComposeExpanding(src)
			if err != nil {
				t.Fatalf("ComposeExpanding failed: %v", err)
			}

			// dst dimensions and pixels must be unchanged.
			if dst.Width() != origDstWidth || dst.Height() != origDstHeight {
				t.Errorf("dst dimensions changed: was %dx%d, now %dx%d",
					origDstWidth, origDstHeight, dst.Width(), dst.Height())
			}
			dstAfter := dst.Render()
			for r := range dstBefore {
				for c := range dstBefore[r] {
					if dstBefore[r][c] != dstAfter[r][c] {
						t.Errorf("dst mutated at (%d,%d): before %v, after %v", r, c, dstBefore[r][c], dstAfter[r][c])
					}
				}
			}

			// Result must have the expected expanded dimensions.
			if result.Width() != tc.wantWidth {
				t.Errorf("result width: want %d, got %d", tc.wantWidth, result.Width())
			}
			if result.Height() != tc.wantHeight {
				t.Errorf("result height: want %d, got %d", tc.wantHeight, result.Height())
			}

			for i, wantFrame := range tc.wantRenderedInOrder {
				gotFrame := result.Render()
				for r := range wantFrame {
					for c := range wantFrame[r] {
						if gotFrame[r][c] != wantFrame[r][c] {
							t.Errorf("frame %d, pixel (%d,%d): expected %v, got %v", i, r, c, wantFrame[r][c], gotFrame[r][c])
						}
					}
				}
			}
		})
	}
}
