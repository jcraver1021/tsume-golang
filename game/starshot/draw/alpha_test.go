package draw_test

import (
	"image/color"
	"testing"

	. "tsumegolang/game/starshot/draw"
)

func TestAlphaComposite(t *testing.T) {
	testCases := []struct {
		name      string
		src       color.RGBA
		dst       color.RGBA
		wantR     uint8
		wantG     uint8
		wantB     uint8
		wantA     uint8
		tolerance uint8 // Allow rounding errors
	}{
		{
			name:  "Opaque over opaque",
			src:   color.RGBA{R: 255, G: 0, B: 0, A: 255}, // Red
			dst:   color.RGBA{R: 0, G: 0, B: 255, A: 255}, // Blue
			wantR: 255, wantG: 0, wantB: 0, wantA: 255,    // Should be red
			tolerance: 1,
		},
		{
			name:  "Transparent over opaque",
			src:   color.RGBA{R: 255, G: 0, B: 0, A: 0},   // Transparent red
			dst:   color.RGBA{R: 0, G: 0, B: 255, A: 255}, // Blue
			wantR: 0, wantG: 0, wantB: 255, wantA: 255,    // Should be blue
			tolerance: 1,
		},
		{
			name:  "Opaque over transparent",
			src:   color.RGBA{R: 255, G: 0, B: 0, A: 255}, // Red
			dst:   color.RGBA{R: 0, G: 0, B: 255, A: 0},   // Transparent blue
			wantR: 255, wantG: 0, wantB: 0, wantA: 255,    // Should be red
			tolerance: 1,
		},
		{
			name:  "50% red over opaque blue",
			src:   color.RGBA{R: 255, G: 0, B: 0, A: 128}, // 50% red
			dst:   color.RGBA{R: 0, G: 0, B: 255, A: 255}, // Blue
			wantR: 127, wantG: 0, wantB: 127, wantA: 255,  // Blend
			tolerance: 2,
		},
		{
			name:  "50% white over opaque black",
			src:   color.RGBA{R: 255, G: 255, B: 255, A: 128}, // 50% white
			dst:   color.RGBA{R: 0, G: 0, B: 0, A: 255},       // Black
			wantR: 127, wantG: 127, wantB: 127, wantA: 255,    // Gray
			tolerance: 2,
		},
		{
			name:  "50% red over 50% blue",
			src:   color.RGBA{R: 255, G: 0, B: 0, A: 128}, // 50% red
			dst:   color.RGBA{R: 0, G: 0, B: 255, A: 128}, // 50% blue
			wantR: 170, wantG: 0, wantB: 85, wantA: 191,   // Blend with alpha
			tolerance: 5, // Higher tolerance for complex blend
		},
		{
			name:  "Transparent over transparent",
			src:   color.RGBA{R: 0, G: 0, B: 0, A: 0},
			dst:   color.RGBA{R: 0, G: 0, B: 0, A: 0},
			wantR: 0, wantG: 0, wantB: 0, wantA: 0,
			tolerance: 0,
		},
		{
			name:  "25% green over opaque red",
			src:   color.RGBA{R: 0, G: 255, B: 0, A: 64},  // 25% green
			dst:   color.RGBA{R: 255, G: 0, B: 0, A: 255}, // Red
			wantR: 191, wantG: 63, wantB: 0, wantA: 255,   // Mostly red with green tint
			tolerance: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			baseMatrix := [][]int{{1}}
			baseColors := map[int]color.RGBA{1: tc.dst}
			base, _ := NewColorMatrix(baseMatrix, baseColors, nil)

			overlayMatrix := [][]int{{2}}
			overlayColors := map[int]color.RGBA{2: tc.src}
			overlay, _ := NewColorMatrix(overlayMatrix, overlayColors, nil)

			base.Compose(overlay, 0, 0)
			result := base.Render()[0][0]

			checkChannel := func(name string, got, want, tol uint8) {
				diff := int(got) - int(want)
				if diff < 0 {
					diff = -diff
				}
				if uint8(diff) > tol {
					t.Errorf("%s = %d, want %d (tolerance %d)", name, got, want, tol)
				}
			}

			checkChannel("R", result.R, tc.wantR, tc.tolerance)
			checkChannel("G", result.G, tc.wantG, tc.tolerance)
			checkChannel("B", result.B, tc.wantB, tc.tolerance)
			checkChannel("A", result.A, tc.wantA, tc.tolerance)
		})
	}
}
