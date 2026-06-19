package environment_test

import (
	"testing"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/environment"
	"tsumegolang/game/starshot/entity/obstacle"
	"tsumegolang/game/starshot/testutil"
)

func TestAsteroidFieldImplementsEntity(t *testing.T) {
	sizeFn := func() obstacle.AsteroidSize { return obstacle.AsteroidMedium }
	field := environment.NewAsteroidField(0.05, sizeFn)

	var _ def.Entity = field

	// Verify all interface methods work
	if got := field.Type(); got != def.EntityTypeEnvironment {
		t.Errorf("Type() = %v, want %v", got, def.EntityTypeEnvironment)
	}

	x, y := field.Location()
	if x != 0 || y != 0 {
		t.Errorf("Location() = (%d,%d), want (0,0)", x, y)
	}

	w, h := field.Dimensions()
	if w != def.ScreenWidth || h != def.ScreenHeight {
		t.Errorf("Dimensions() = (%d,%d), want (%d,%d)", w, h, def.ScreenWidth, def.ScreenHeight)
	}

	if field.CanBeRemoved() {
		t.Error("AsteroidField should never be removed")
	}

	if field.BoundingBoxOverlaps(field) {
		t.Error("AsteroidField should not overlap anything")
	}
}

func TestAsteroidFieldGeneratesAsteroids(t *testing.T) {
	scene := testutil.NewMockScene()
	sizeFn := func() obstacle.AsteroidSize { return obstacle.AsteroidMedium }
	field := environment.NewAsteroidField(0.5, sizeFn) // High density for reliable test

	initialCount := scene.EntityCount()

	// Call Act() multiple times
	for range 100 {
		field.Act(scene)
	}

	got := scene.EntityCount()

	if got <= initialCount {
		t.Errorf("asteroids generated = %d, want > %d", got, initialCount)
	}

	t.Logf("Generated %d asteroids in 100 ticks", got-initialCount)
}

func TestAsteroidFieldDensityAffectsGeneration(t *testing.T) {
	iterations := 200
	sizeFn := func() obstacle.AsteroidSize { return obstacle.AsteroidSmall }

	// Low density
	sceneLow := testutil.NewMockScene()
	fieldLow := environment.NewAsteroidField(0.01, sizeFn)
	for i := 0; i < iterations; i++ {
		fieldLow.Act(sceneLow)
	}

	// High density
	sceneHigh := testutil.NewMockScene()
	fieldHigh := environment.NewAsteroidField(0.10, sizeFn)
	for i := 0; i < iterations; i++ {
		fieldHigh.Act(sceneHigh)
	}

	got := sceneHigh.EntityCount()
	baseline := sceneLow.EntityCount()

	if got <= baseline {
		t.Errorf("high density asteroids = %d, want > %d (low density)", got, baseline)
	}

	t.Logf("Low density (0.01): %d asteroids, High density (0.10): %d asteroids", baseline, got)
}

func TestAsteroidFieldUsesSizeFunction(t *testing.T) {
	scene := testutil.NewMockScene()

	// Test with fixed size function
	expectedSize := obstacle.AsteroidLarge
	sizeFn := func() obstacle.AsteroidSize { return expectedSize }
	field := environment.NewAsteroidField(1.0, sizeFn) // 100% density for deterministic test

	// Generate one asteroid
	field.Act(scene)

	if scene.EntityCount() == 0 {
		t.Fatal("No asteroids generated")
	}

	// Check that asteroid has the expected dimensions
	asteroid := scene.GetEntities()[0]
	expectedWidth, expectedHeight := expectedSize.Dimensions()
	gotWidth, gotHeight := asteroid.Dimensions()

	if gotWidth != expectedWidth || gotHeight != expectedHeight {
		t.Errorf("Asteroid dimensions = (%d,%d), want (%d,%d) for size %v",
			gotWidth, gotHeight, expectedWidth, expectedHeight, expectedSize)
	}
}

func TestAsteroidFieldRandomSizeDistribution(t *testing.T) {
	scene := testutil.NewMockScene()

	// Size function that alternates between two sizes
	count := 0
	sizeFn := func() obstacle.AsteroidSize {
		count++
		if count%2 == 0 {
			return obstacle.AsteroidSmall
		}
		return obstacle.AsteroidLarge
	}

	field := environment.NewAsteroidField(1.0, sizeFn)

	// Generate multiple asteroids
	for i := 0; i < 10; i++ {
		field.Act(scene)
	}

	// Count asteroids by size
	smallWidth, _ := obstacle.AsteroidSmall.Dimensions()
	largeWidth, _ := obstacle.AsteroidLarge.Dimensions()

	smallCount := 0
	largeCount := 0

	for _, entity := range scene.GetEntities() {
		w, _ := entity.Dimensions()
		switch w {
		case smallWidth:
			smallCount++
		case largeWidth:
			largeCount++
		}
	}

	if smallCount == 0 {
		t.Error("No small asteroids generated")
	}
	if largeCount == 0 {
		t.Error("No large asteroids generated")
	}

	t.Logf("Generated: Small=%d, Large=%d", smallCount, largeCount)
}

func TestAsteroidFieldStartsAboveScreen(t *testing.T) {
	scene := testutil.NewMockScene()
	sizeFn := func() obstacle.AsteroidSize { return obstacle.AsteroidMedium }
	field := environment.NewAsteroidField(1.0, sizeFn) // 100% density

	// Generate asteroids
	for i := 0; i < 50; i++ {
		field.Act(scene)
	}

	// All asteroids should start above screen (y <= 0)
	for i, entity := range scene.GetEntities() {
		if entity.Type() != def.EntityTypeObstacle {
			continue
		}

		_, y := entity.Location()
		if y > 0 {
			t.Errorf("Asteroid %d at y=%d, want <= 0 to avoid pop-in", i, y)
		}
	}
}

func TestAsteroidFieldGeneratesObstacleType(t *testing.T) {
	scene := testutil.NewMockScene()
	sizeFn := func() obstacle.AsteroidSize { return obstacle.AsteroidMedium }
	field := environment.NewAsteroidField(1.0, sizeFn)

	field.Act(scene)

	if scene.EntityCount() == 0 {
		t.Fatal("No asteroids generated")
	}

	asteroid := scene.GetEntities()[0]
	if got := asteroid.Type(); got != def.EntityTypeObstacle {
		t.Errorf("Generated entity type = %v, want %v", got, def.EntityTypeObstacle)
	}
}

func TestAsteroidFieldWithZeroDensity(t *testing.T) {
	scene := testutil.NewMockScene()
	sizeFn := func() obstacle.AsteroidSize { return obstacle.AsteroidSmall }
	field := environment.NewAsteroidField(0.0, sizeFn) // 0% density

	// Should generate very few or no asteroids
	for i := 0; i < 100; i++ {
		field.Act(scene)
	}

	got := scene.EntityCount()
	// With 0 density, should generate 0 asteroids (or extremely rare due to randomness)
	if got > 5 {
		t.Errorf("Zero density generated %d asteroids, want ~0", got)
	}
}

func TestAsteroidFieldPositionDistribution(t *testing.T) {
	scene := testutil.NewMockScene()
	sizeFn := func() obstacle.AsteroidSize { return obstacle.AsteroidTiny }
	field := environment.NewAsteroidField(1.0, sizeFn)

	// Generate many asteroids
	for range 100 {
		field.Act(scene)
	}

	// Check X position distribution (should cover full width)
	minX := scene.Width()
	maxX := 0

	for _, entity := range scene.GetEntities() {
		if entity.Type() != def.EntityTypeObstacle {
			continue
		}
		x, _ := entity.Location()
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
	}

	// Should use at least 80% of screen width
	got := maxX - minX
	want := int(float64(scene.Width()-1) * 0.8)

	if got < want {
		t.Errorf("X range = %d, want >= %d (80%% of screen width)", got, want)
	}

	t.Logf("Asteroid X range: %d to %d (range=%d)", minX, maxX, got)
}
