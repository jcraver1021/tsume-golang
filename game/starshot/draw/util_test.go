package draw_test

import (
	"testing"

	. "tsumegolang/game/starshot/draw"
)

func TestNewMatrix(t *testing.T) {
	testCases := []struct {
		name   string
		width  int
		height int
	}{
		{"3x3", 3, 3},
		{"5x2", 5, 2},
		{"0x0", 0, 0},
		{"1x1", 1, 1},
		{"10x5", 10, 5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matrix := NewMatrix(tc.width, tc.height)
			if len(matrix) != tc.height {
				t.Errorf("Expected height %d, got %d", tc.height, len(matrix))
			}
			for _, row := range matrix {
				if len(row) != tc.width {
					t.Errorf("Expected width %d, got %d", tc.width, len(row))
				}
			}
		})
	}
}
