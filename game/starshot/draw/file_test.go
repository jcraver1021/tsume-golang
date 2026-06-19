package draw_test

import (
	"image/color"
	"os"
	"path/filepath"
	"testing"

	. "tsumegolang/game/starshot/draw"
)

func TestColorMatrixFromFile(t *testing.T) {
	// Create a temporary YAML file
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "test_matrix.yaml")

	yamlContent := `matrix:
  - [1, 2, 3]
  - [2, 3, 4]
  - [3, 4, 5]

color_codes:
  1: "#FF0000FF"  # Red
  2: "#00FF00FF"  # Green
  3: "#0000FFFF"  # Blue

animation_sequences:
  4:
    frames:
      - "#FFFF00FF"  # Yellow
      - "#FF00FFFF"  # Magenta
    frame_duration: 5
  5:
    frames:
      - "#00FFFFFF"  # Cyan
      - "#FFFFFFFF"  # White
    frame_duration: 10
`

	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}

	// Load the color matrix
	cm, err := ColorMatrixFromFile(yamlPath)
	if err != nil {
		t.Fatalf("ColorMatrixFromFile() error = %v", err)
	}

	// Verify matrix dimensions
	if len(cm.Matrix) != 3 {
		t.Errorf("Matrix height = %d, want 3", len(cm.Matrix))
	}
	if len(cm.Matrix[0]) != 3 {
		t.Errorf("Matrix width = %d, want 3", len(cm.Matrix[0]))
	}

	// Verify matrix contents
	want := [][]int{
		{1, 2, 3},
		{2, 3, 4},
		{3, 4, 5},
	}
	for i := range cm.Matrix {
		for j := range cm.Matrix[i] {
			if cm.Matrix[i][j] != want[i][j] {
				t.Errorf("Matrix[%d][%d] = %d, want %d", i, j, cm.Matrix[i][j], want[i][j])
			}
		}
	}

	// Verify color codes were converted correctly
	expectedColors := map[int]color.RGBA{
		1: {R: 255, G: 0, B: 0, A: 255}, // Red
		2: {R: 0, G: 255, B: 0, A: 255}, // Green
		3: {R: 0, G: 0, B: 255, A: 255}, // Blue
	}

	for code, expectedColor := range expectedColors {
		gotColor, exists := cm.ColorCodes[code]
		if !exists {
			t.Errorf("ColorCodes[%d] does not exist", code)
			continue
		}
		if gotColor != expectedColor {
			t.Errorf("ColorCodes[%d] = %+v, want %+v", code, gotColor, expectedColor)
		}
	}
}

func TestColorMatrixFromFileInvalidPath(t *testing.T) {
	_, err := ColorMatrixFromFile("/nonexistent/path/to/file.yaml")
	if err == nil {
		t.Error("ColorMatrixFromFile() with invalid path should return error")
	}
}

func TestColorMatrixFromFileInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	invalidYAML := `matrix: [this is not valid YAML syntax`
	if err := os.WriteFile(yamlPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := ColorMatrixFromFile(yamlPath)
	if err == nil {
		t.Error("ColorMatrixFromFile() with invalid YAML should return error")
	}
}

func TestColorMatrixFromFileInvalidHexColor(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "invalid_color.yaml")

	yamlContent := `matrix:
  - [1, 2]

color_codes:
  1: "#INVALID"  # Invalid hex format
  2: "#00FF00FF"
`

	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := ColorMatrixFromFile(yamlPath)
	if err == nil {
		t.Error("ColorMatrixFromFile() with invalid hex color should return error")
	}
}

func TestColorMatrixFromFileWithAnimations(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "animated.yaml")

	yamlContent := `matrix:
  - [1, 2]
  - [2, 3]

color_codes:
  1: "#FF0000FF"

animation_sequences:
  2:
    frames:
      - "#00FF00FF"
      - "#0000FFFF"
    frame_duration: 4
  3:
    frames:
      - "#FFFF00FF"
    frame_duration: 1
`

	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	cm, err := ColorMatrixFromFile(yamlPath)
	if err != nil {
		t.Fatalf("ColorMatrixFromFile() error = %v", err)
	}

	// Verify we have 1 static color and 2 animations
	if len(cm.ColorCodes) != 1 {
		t.Errorf("ColorCodes length = %d, want 1", len(cm.ColorCodes))
	}

	// Render to verify animations work
	rendered := cm.Render()
	if len(rendered) != 2 || len(rendered[0]) != 2 {
		t.Errorf("Rendered dimensions = %dx%d, want 2x2", len(rendered), len(rendered[0]))
	}

	// First render should show first frame of animations
	// Code 2 should be green (first frame)
	expectedGreen := color.RGBA{R: 0, G: 255, B: 0, A: 255}
	if rendered[0][1] != expectedGreen {
		t.Errorf("rendered[0][1] = %+v, want %+v (first frame of animation)", rendered[0][1], expectedGreen)
	}
}
