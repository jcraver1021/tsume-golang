package obstacle_test

import (
	"testing"

	"tsumegolang/game/starshot/entity/obstacle"
)

func TestAllAsteroidSizes(t *testing.T) {
	tests := []struct {
		name       string
		size       obstacle.AsteroidSize
		wantWidth  int
		wantHeight int
		wantSpeed  int
	}{
		{"tiny", obstacle.AsteroidTiny, 8, 8, 4},
		{"small", obstacle.AsteroidSmall, 12, 12, 3},
		{"medium", obstacle.AsteroidMedium, 20, 20, 2},
		{"large", obstacle.AsteroidLarge, 32, 32, 2},
		{"huge", obstacle.AsteroidHuge, 48, 48, 1},
		{"massive", obstacle.AsteroidMassive, 64, 64, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asteroid := obstacle.NewAsteroid(100, 200, tt.size)

			gotWidth, gotHeight := asteroid.Dimensions()
			if gotWidth != tt.wantWidth || gotHeight != tt.wantHeight {
				t.Errorf("Size %s: Dimensions() = (%d, %d), want (%d, %d)",
					tt.name, gotWidth, gotHeight, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestNewRandomAsteroidInRange(t *testing.T) {
	tests := []struct {
		name       string
		minSize    obstacle.AsteroidSize
		maxSize    obstacle.AsteroidSize
		wantMinDim int
		wantMaxDim int
	}{
		{
			name:       "tiny_to_small",
			minSize:    obstacle.AsteroidTiny,
			maxSize:    obstacle.AsteroidSmall,
			wantMinDim: 8,
			wantMaxDim: 12,
		},
		{
			name:       "large_to_massive",
			minSize:    obstacle.AsteroidLarge,
			maxSize:    obstacle.AsteroidMassive,
			wantMinDim: 32,
			wantMaxDim: 64,
		},
		{
			name:       "only_huge",
			minSize:    obstacle.AsteroidHuge,
			maxSize:    obstacle.AsteroidHuge,
			wantMinDim: 48,
			wantMaxDim: 48,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate multiple asteroids to test range
			seenSizes := make(map[int]bool)

			for range 20 {
				asteroid := obstacle.NewRandomAsteroidInRange(0, 0, tt.minSize, tt.maxSize)
				width, _ := asteroid.Dimensions()

				if width < tt.wantMinDim || width > tt.wantMaxDim {
					t.Errorf("Asteroid dimension %d outside range [%d, %d]",
						width, tt.wantMinDim, tt.wantMaxDim)
				}

				seenSizes[width] = true
			}

			// If range has multiple sizes, we should see variation
			if tt.minSize != tt.maxSize && len(seenSizes) < 2 {
				t.Logf("Warning: Only saw %d different sizes in 20 iterations (expected variety)", len(seenSizes))
			}
		})
	}
}

func TestMassiveAsteroidSize(t *testing.T) {
	// Massive asteroids should be 64x64
	asteroid := obstacle.NewAsteroid(0, 0, obstacle.AsteroidMassive)

	width, height := asteroid.Dimensions()
	if width != 64 || height != 64 {
		t.Errorf("Massive asteroid dimensions = (%d, %d), want (64, 64)", width, height)
	}

	// Should be slow
	asteroid.Act(nil)
	_, newY := asteroid.Location()
	if newY != 1 {
		t.Errorf("Massive asteroid moved to y=%d, want 1 (speed should be 1)", newY)
	}
}

func TestTinyAsteroidSize(t *testing.T) {
	// Tiny asteroids should be 8x8
	asteroid := obstacle.NewAsteroid(0, 0, obstacle.AsteroidTiny)

	width, height := asteroid.Dimensions()
	if width != 8 || height != 8 {
		t.Errorf("Tiny asteroid dimensions = (%d, %d), want (8, 8)", width, height)
	}

	// Should be fast
	asteroid.Act(nil)
	_, newY := asteroid.Location()
	if newY != 4 {
		t.Errorf("Tiny asteroid moved to y=%d, want 4 (speed should be 4)", newY)
	}
}
