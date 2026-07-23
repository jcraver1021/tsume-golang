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
	testCases := []struct {
		name       string
		size       obstacle.AsteroidSize
		wantWidth  int
		wantHeight int
	}{
		{"tiny", obstacle.AsteroidTiny, 8, 8},
		{"small", obstacle.AsteroidSmall, 12, 12},
		{"medium", obstacle.AsteroidMedium, 20, 20},
		{"large", obstacle.AsteroidLarge, 32, 32},
		{"huge", obstacle.AsteroidHuge, 48, 48},
		{"massive", obstacle.AsteroidMassive, 64, 64},
		{"gigantic", obstacle.AsteroidGigantic, 80, 80},
		{"colossal", obstacle.AsteroidColossal, 96, 96},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			asteroid := obstacle.NewAsteroid(100, 200, tc.size)

			gotX, gotY := asteroid.Location()
			if gotX != 100 || gotY != 200 {
				t.Errorf("Location() = (%d, %d), want (100, 200)", gotX, gotY)
			}

			gotWidth, gotHeight := asteroid.Dimensions()
			if gotWidth != tc.wantWidth || gotHeight != tc.wantHeight {
				t.Errorf("Dimensions() = (%d, %d), want (%d, %d)",
					gotWidth, gotHeight, tc.wantWidth, tc.wantHeight)
			}
		})
	}
}

func TestAsteroidMovement(t *testing.T) {
	scene := testutil.NewMockScene()
	asteroid := obstacle.NewAsteroid(100, 100, obstacle.AsteroidSmall)

	asteroid.Act(scene)

	gotX, gotY := asteroid.Location()
	if gotX != 100 || gotY != 103 {
		t.Errorf("After Act(), Location() = (%d, %d), want (100, 103)", gotX, gotY)
	}
}

func TestAsteroidRemoval(t *testing.T) {
	testCases := []struct {
		name string
		y    int
		want bool
	}{
		{"on screen", 100, false},
		{"at bottom", def.ScreenHeight, false},
		{"past bottom", def.ScreenHeight + 1, true},
		{"far past bottom", def.ScreenHeight + 100, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			asteroid := obstacle.NewAsteroid(100, tc.y, obstacle.AsteroidSmall)
			if got := asteroid.CanBeRemoved(); got != tc.want {
				t.Errorf("CanBeRemoved() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAsteroidBoundingBoxOverlaps(t *testing.T) {
	asteroid := obstacle.NewAsteroid(100, 100, obstacle.AsteroidSmall)
	other := obstacle.NewAsteroid(105, 105, obstacle.AsteroidSmall)

	if !asteroid.BoundingBoxOverlaps(other) {
		t.Error("BoundingBoxOverlaps() = false, want true for overlapping asteroids")
	}

	farAway := obstacle.NewAsteroid(200, 200, obstacle.AsteroidSmall)
	if asteroid.BoundingBoxOverlaps(farAway) {
		t.Error("BoundingBoxOverlaps() = true, want false for non-overlapping asteroids")
	}
}

func TestNewRandomAsteroid(t *testing.T) {
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

		sizes[width] = true
	}

	if len(sizes) < 2 {
		t.Errorf("Only saw %d different sizes, expected at least 2 (random distribution)", len(sizes))
	}
}

func TestNewRandomAsteroidInRange(t *testing.T) {
	testCases := []struct {
		name       string
		minSize    obstacle.AsteroidSize
		maxSize    obstacle.AsteroidSize
		wantMinDim int
		wantMaxDim int
	}{
		{"tiny_to_small", obstacle.AsteroidTiny, obstacle.AsteroidSmall, 8, 12},
		{"large_to_massive", obstacle.AsteroidLarge, obstacle.AsteroidMassive, 32, 64},
		{"only_huge", obstacle.AsteroidHuge, obstacle.AsteroidHuge, 48, 48},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			seenSizes := make(map[int]bool)

			for range 20 {
				asteroid := obstacle.NewRandomAsteroidInRange(0, 0, tc.minSize, tc.maxSize)
				width, _ := asteroid.Dimensions()

				if width < tc.wantMinDim || width > tc.wantMaxDim {
					t.Errorf("Asteroid dimension %d outside range [%d, %d]",
						width, tc.wantMinDim, tc.wantMaxDim)
				}

				seenSizes[width] = true
			}

			if tc.minSize != tc.maxSize && len(seenSizes) < 2 {
				t.Logf("Warning: Only saw %d different sizes in 20 iterations (expected variety)", len(seenSizes))
			}
		})
	}
}

func TestAsteroidProceduralVariation(t *testing.T) {
	a1 := obstacle.NewAsteroid(0, 0, obstacle.AsteroidSmall)
	a2 := obstacle.NewAsteroid(0, 0, obstacle.AsteroidSmall)

	if a1 == nil || a2 == nil {
		t.Fatal("Failed to create asteroids")
	}

	w1, h1 := a1.Dimensions()
	w2, h2 := a2.Dimensions()

	if w1 != 12 || h1 != 12 || w2 != 12 || h2 != 12 {
		t.Errorf("Small asteroids should be 12x12, got (%d,%d) and (%d,%d)", w1, h1, w2, h2)
	}
}
