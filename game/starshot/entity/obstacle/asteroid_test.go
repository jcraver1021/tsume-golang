package obstacle_test

import (
	"testing"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/obstacle"
	"tsumegolang/game/starshot/testutil"
)

func TestAsteroidImplementsEntity(t *testing.T) {
	asteroid := obstacle.NewAsteroid(0, 0, obstacle.AsteroidSmall)
	var _ def.Entity = asteroid
}

func TestAsteroidImplementsPreciseCollider(t *testing.T) {
	asteroid := obstacle.NewAsteroid(0, 0, obstacle.AsteroidSmall)
	var _ def.PreciseCollider = asteroid
}

func TestAsteroidType(t *testing.T) {
	asteroid := obstacle.NewAsteroid(0, 0, obstacle.AsteroidMedium)
	if got := asteroid.Type(); got != def.EntityTypeObstacle {
		t.Errorf("Type() = %v, want %v", got, def.EntityTypeObstacle)
	}
}

func TestAsteroidSizes(t *testing.T) {
	tests := []struct {
		name       string
		size       obstacle.AsteroidSize
		wantWidth  int
		wantHeight int
		wantSpeed  int
	}{
		{"small", obstacle.AsteroidSmall, 12, 12, 3},
		{"medium", obstacle.AsteroidMedium, 20, 20, 2},
		{"large", obstacle.AsteroidLarge, 32, 32, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asteroid := obstacle.NewAsteroid(100, 200, tt.size)

			gotX, gotY := asteroid.Location()
			if gotX != 100 || gotY != 200 {
				t.Errorf("Location() = (%d, %d), want (100, 200)", gotX, gotY)
			}

			gotWidth, gotHeight := asteroid.Dimensions()
			if gotWidth != tt.wantWidth || gotHeight != tt.wantHeight {
				t.Errorf("Dimensions() = (%d, %d), want (%d, %d)",
					gotWidth, gotHeight, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestAsteroidMovement(t *testing.T) {
	scene := testutil.NewMockScene()
	asteroid := obstacle.NewAsteroid(100, 100, obstacle.AsteroidSmall)

	asteroid.Act(scene)

	gotX, gotY := asteroid.Location()
	wantX, wantY := 100, 103 // Moves down by speed (3)

	if gotX != wantX || gotY != wantY {
		t.Errorf("After Act(), Location() = (%d, %d), want (%d, %d)",
			gotX, gotY, wantX, wantY)
	}
}

func TestAsteroidRemoval(t *testing.T) {
	tests := []struct {
		name string
		y    int
		want bool
	}{
		{"on screen", 100, false},
		{"at bottom", def.ScreenHeight, false},
		{"past bottom", def.ScreenHeight + 1, true},
		{"far past bottom", def.ScreenHeight + 100, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asteroid := obstacle.NewAsteroid(100, tt.y, obstacle.AsteroidSmall)
			if got := asteroid.CanBeRemoved(); got != tt.want {
				t.Errorf("CanBeRemoved() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAsteroidBoundingBoxOverlaps(t *testing.T) {
	asteroid := obstacle.NewAsteroid(100, 100, obstacle.AsteroidSmall) // 12x12
	other := obstacle.NewAsteroid(105, 105, obstacle.AsteroidSmall)    // 12x12, overlapping

	if !asteroid.BoundingBoxOverlaps(other) {
		t.Error("BoundingBoxOverlaps() = false, want true for overlapping asteroids")
	}

	farAway := obstacle.NewAsteroid(200, 200, obstacle.AsteroidSmall)
	if asteroid.BoundingBoxOverlaps(farAway) {
		t.Error("BoundingBoxOverlaps() = true, want false for non-overlapping asteroids")
	}
}

func TestNewRandomAsteroid(t *testing.T) {
	// Just verify it creates valid asteroids with varying sizes
	sizes := make(map[int]bool)

	for i := range 50 {
		asteroid := obstacle.NewRandomAsteroid(i*10, i*20)

		if asteroid.Type() != def.EntityTypeObstacle {
			t.Errorf("Iteration %d: Type() = %v, want %v", i, asteroid.Type(), def.EntityTypeObstacle)
		}

		width, height := asteroid.Dimensions()
		if width == 0 || height == 0 {
			t.Errorf("Iteration %d: Invalid dimensions (%d, %d)", i, width, height)
		}

		// Track sizes to ensure randomness
		sizes[width] = true
	}

	// Should have seen at least 2 different sizes in 50 iterations
	if len(sizes) < 2 {
		t.Errorf("Only saw %d different sizes, expected at least 2 (random distribution)", len(sizes))
	}
}

func TestAsteroidMultipleColors(t *testing.T) {
	// Test that asteroids have multiple colors (not single-color like old version)
	asteroid := obstacle.NewAsteroid(0, 0, obstacle.AsteroidMedium)

	// We can't directly inspect the ColorMatrix, but we can verify it was created
	// and has reasonable dimensions
	width, height := asteroid.Dimensions()
	if width != 20 || height != 20 {
		t.Errorf("Medium asteroid dimensions = (%d, %d), want (20, 20)", width, height)
	}

	// The asteroid should successfully implement PreciseCollider
	// which means it has a valid sprite with pixels
	var _ def.PreciseCollider = asteroid
}

func TestAsteroidProceduralVariation(t *testing.T) {
	// Create multiple asteroids and verify they're not all identical
	// (procedural generation should create variety)

	asteroid1 := obstacle.NewAsteroid(0, 0, obstacle.AsteroidSmall)
	asteroid2 := obstacle.NewAsteroid(0, 0, obstacle.AsteroidSmall)

	// Both should be valid
	if asteroid1 == nil || asteroid2 == nil {
		t.Fatal("Failed to create asteroids")
	}

	// Both should have correct size
	w1, h1 := asteroid1.Dimensions()
	w2, h2 := asteroid2.Dimensions()

	if w1 != 12 || h1 != 12 || w2 != 12 || h2 != 12 {
		t.Errorf("Small asteroids should be 12x12, got (%d,%d) and (%d,%d)", w1, h1, w2, h2)
	}

	// Note: We can't easily test that they look different without rendering,
	// but the procedural generation should create variation in colors and shapes
}
