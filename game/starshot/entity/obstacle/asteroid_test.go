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
		{"small", obstacle.AsteroidSmall, 8, 8, 3},
		{"medium", obstacle.AsteroidMedium, 16, 16, 2},
		{"large", obstacle.AsteroidLarge, 24, 24, 1},
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
	asteroid := obstacle.NewAsteroid(100, 100, obstacle.AsteroidSmall) // 8x8
	other := obstacle.NewAsteroid(105, 105, obstacle.AsteroidSmall)     // 8x8, overlapping

	if !asteroid.BoundingBoxOverlaps(other) {
		t.Error("BoundingBoxOverlaps() = false, want true for overlapping asteroids")
	}

	farAway := obstacle.NewAsteroid(200, 200, obstacle.AsteroidSmall)
	if asteroid.BoundingBoxOverlaps(farAway) {
		t.Error("BoundingBoxOverlaps() = true, want false for non-overlapping asteroids")
	}
}

func TestNewRandomAsteroid(t *testing.T) {
	// Just verify it creates valid asteroids
	for i := range 10 {
		asteroid := obstacle.NewRandomAsteroid(i*10, i*20)

		if asteroid.Type() != def.EntityTypeObstacle {
			t.Errorf("Iteration %d: Type() = %v, want %v", i, asteroid.Type(), def.EntityTypeObstacle)
		}

		width, height := asteroid.Dimensions()
		if width == 0 || height == 0 {
			t.Errorf("Iteration %d: Invalid dimensions (%d, %d)", i, width, height)
		}
	}
}
