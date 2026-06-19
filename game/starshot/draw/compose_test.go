package draw_test

import (
	"image/color"
	"testing"

	. "tsumegolang/game/starshot/draw"
)

func TestComposeSkipsTransparentPixels(t *testing.T) {
	// Create a base matrix with a solid color
	baseMatrix := [][]int{
		{1, 1, 1},
		{1, 1, 1},
		{1, 1, 1},
	}
	baseColors := map[int]color.RGBA{
		1: {R: 255, G: 0, B: 0, A: 255}, // Red
	}

	base, err := NewColorMatrix(baseMatrix, baseColors, nil)
	if err != nil {
		t.Fatalf("Failed to create base matrix: %v", err)
	}

	// Create overlay with transparent pixels and one opaque pixel
	overlayMatrix := [][]int{
		{0, 0, 0},
		{0, 2, 0},
		{0, 0, 0},
	}
	overlayColors := map[int]color.RGBA{
		0: {R: 0, G: 0, B: 0, A: 0},     // Transparent
		2: {R: 0, G: 255, B: 0, A: 255}, // Green
	}

	overlay, err := NewColorMatrix(overlayMatrix, overlayColors, nil)
	if err != nil {
		t.Fatalf("Failed to create overlay matrix: %v", err)
	}

	// Compose overlay onto base
	if err := base.Compose(overlay, 0, 0); err != nil {
		t.Fatalf("Compose failed: %v", err)
	}

	// Render and check results
	rendered := base.Render()

	// Center pixel should be green (overlay)
	if rendered[1][1] != (color.RGBA{R: 0, G: 255, B: 0, A: 255}) {
		t.Errorf("Center pixel = %+v, want green", rendered[1][1])
	}

	// Corner pixels should still be red (base not overwritten by transparent)
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if rendered[0][0] != red {
		t.Errorf("Corner pixel = %+v, want red (transparent should not overwrite)", rendered[0][0])
	}
	if rendered[2][2] != red {
		t.Errorf("Corner pixel = %+v, want red (transparent should not overwrite)", rendered[2][2])
	}
}

func TestComposeWithAnimatedTransparentPixels(t *testing.T) {
	// Base matrix with solid color
	baseMatrix := [][]int{{1, 1}, {1, 1}}
	baseColors := map[int]color.RGBA{
		1: {R: 255, G: 0, B: 0, A: 255}, // Red
	}

	base, err := NewColorMatrix(baseMatrix, baseColors, nil)
	if err != nil {
		t.Fatalf("Failed to create base matrix: %v", err)
	}

	// Overlay with transparent animation and one opaque pixel
	overlayMatrix := [][]int{{0, 2}, {0, 0}}
	overlayColors := map[int]color.RGBA{
		2: {R: 0, G: 255, B: 0, A: 255}, // Green
	}
	overlayAnimations := map[int]*AnimationSequence{
		0: NewAnimationSequence([]color.RGBA{
			{R: 0, G: 0, B: 0, A: 0}, // Transparent
			{R: 0, G: 0, B: 0, A: 0}, // Transparent
		}, 1),
	}

	overlay, err := NewColorMatrix(overlayMatrix, overlayColors, overlayAnimations)
	if err != nil {
		t.Fatalf("Failed to create overlay matrix: %v", err)
	}

	// Compose
	if err := base.Compose(overlay, 0, 0); err != nil {
		t.Fatalf("Compose failed: %v", err)
	}

	// Render
	rendered := base.Render()

	// Top-right should be green (opaque overlay)
	if rendered[0][1] != (color.RGBA{R: 0, G: 255, B: 0, A: 255}) {
		t.Errorf("Top-right pixel = %+v, want green", rendered[0][1])
	}

	// Other pixels should still be red (transparent animation should not overwrite)
	red := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if rendered[0][0] != red {
		t.Errorf("Top-left pixel = %+v, want red (transparent animation should not overwrite)", rendered[0][0])
	}
}

func TestComposeSemiTransparentBlending(t *testing.T) {
	// Base: solid red
	baseMatrix := [][]int{{1}}
	baseColors := map[int]color.RGBA{
		1: {R: 255, G: 0, B: 0, A: 255}, // Opaque red
	}

	base, err := NewColorMatrix(baseMatrix, baseColors, nil)
	if err != nil {
		t.Fatalf("Failed to create base: %v", err)
	}

	// Overlay: 50% transparent green
	overlayMatrix := [][]int{{2}}
	overlayColors := map[int]color.RGBA{
		2: {R: 0, G: 255, B: 0, A: 128}, // 50% green
	}

	overlay, err := NewColorMatrix(overlayMatrix, overlayColors, nil)
	if err != nil {
		t.Fatalf("Failed to create overlay: %v", err)
	}

	// Compose
	if err := base.Compose(overlay, 0, 0); err != nil {
		t.Fatalf("Compose failed: %v", err)
	}

	// Render
	rendered := base.Render()
	result := rendered[0][0]

	// Should be a blend: some red + some green
	// With 50% alpha green over opaque red:
	// R = (0 * 0.5 + 255 * 0.5) / 1.0 = 127
	// G = (255 * 0.5 + 0 * 0.5) / 1.0 = 127
	// B = 0
	// A = 0.5 + 1.0 * (1 - 0.5) = 1.0 = 255

	if result.A != 255 {
		t.Errorf("Blended alpha = %d, want 255 (fully opaque)", result.A)
	}

	// RGB should be around 127 each (allowing some rounding error)
	if result.R < 120 || result.R > 135 {
		t.Errorf("Blended R = %d, want ~127", result.R)
	}
	if result.G < 120 || result.G > 135 {
		t.Errorf("Blended G = %d, want ~127", result.G)
	}
	if result.B != 0 {
		t.Errorf("Blended B = %d, want 0", result.B)
	}
}

func TestComposeMultipleSemiTransparent(t *testing.T) {
	// Base: solid white
	baseMatrix := [][]int{{1, 1}, {1, 1}}
	baseColors := map[int]color.RGBA{
		1: {R: 255, G: 255, B: 255, A: 255}, // White
	}

	base, err := NewColorMatrix(baseMatrix, baseColors, nil)
	if err != nil {
		t.Fatalf("Failed to create base: %v", err)
	}

	// First overlay: 50% red top-left
	overlay1Matrix := [][]int{{2, 0}, {0, 0}}
	overlay1Colors := map[int]color.RGBA{
		0: {R: 0, G: 0, B: 0, A: 0},     // Transparent
		2: {R: 255, G: 0, B: 0, A: 128}, // 50% red
	}

	overlay1, err := NewColorMatrix(overlay1Matrix, overlay1Colors, nil)
	if err != nil {
		t.Fatalf("Failed to create overlay1: %v", err)
	}

	// Second overlay: 50% blue bottom-right
	overlay2Matrix := [][]int{{0, 0}, {0, 3}}
	overlay2Colors := map[int]color.RGBA{
		0: {R: 0, G: 0, B: 0, A: 0},     // Transparent
		3: {R: 0, G: 0, B: 255, A: 128}, // 50% blue
	}

	overlay2, err := NewColorMatrix(overlay2Matrix, overlay2Colors, nil)
	if err != nil {
		t.Fatalf("Failed to create overlay2: %v", err)
	}

	// Compose both
	if err := base.Compose(overlay1, 0, 0); err != nil {
		t.Fatalf("Compose overlay1 failed: %v", err)
	}
	if err := base.Compose(overlay2, 0, 0); err != nil {
		t.Fatalf("Compose overlay2 failed: %v", err)
	}

	rendered := base.Render()

	// Top-left: white + 50% red = pinkish
	topLeft := rendered[0][0]
	if topLeft.R < 250 { // Should be high red
		t.Errorf("Top-left R = %d, should be high (red tint)", topLeft.R)
	}

	// Bottom-right: white + 50% blue = light blue
	bottomRight := rendered[1][1]
	if bottomRight.B < 250 { // Should be high blue
		t.Errorf("Bottom-right B = %d, should be high (blue tint)", bottomRight.B)
	}

	// Top-right: unchanged white
	topRight := rendered[0][1]
	if topRight != (color.RGBA{R: 255, G: 255, B: 255, A: 255}) {
		t.Errorf("Top-right = %+v, should still be white", topRight)
	}
}
