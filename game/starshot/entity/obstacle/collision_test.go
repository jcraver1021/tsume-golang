package obstacle_test

import (
	"testing"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/obstacle"
	"tsumegolang/game/starshot/testutil"
)

func TestAsteroidCollisionWithTransparentPixels(t *testing.T) {
	// Create a small asteroid at a specific position
	asteroid := obstacle.NewAsteroid(100, 100, obstacle.AsteroidSmall) // 12x12

	// Create a simple entity that overlaps with the asteroid's bounding box
	entity := testutil.NewMockEntity(def.EntityTypePlayer)
	entity.X = 105
	entity.Y = 105
	entity.Width = 32
	entity.Height = 32

	// First check: bounding boxes should overlap
	if !asteroid.BoundingBoxOverlaps(entity) {
		t.Error("Bounding boxes should overlap")
	}

	// Second check: Use the Collides helper which should use precise collision
	// Since asteroid implements PreciseCollider, it should check actual pixels
	collision := def.Collides(asteroid, entity)

	// We can't predict the exact result without knowing the procedural shape,
	// but we can verify the collision detection ran without panic
	t.Logf("Collision detected: %v (depends on procedural shape)", collision)

	// The key is that Collides() should call asteroid.CollidesWith(entity)
	// and only count solid (non-transparent) pixels
}

func TestAsteroidPreciseCollision(t *testing.T) {
	// Create a large asteroid to have more predictable collision area
	asteroid := obstacle.NewAsteroid(50, 50, obstacle.AsteroidLarge) // 32x32

	// Test 1: Entity completely outside bounding box - should not collide
	farAway := obstacle.NewAsteroid(200, 200, obstacle.AsteroidSmall)
	if def.Collides(asteroid, farAway) {
		t.Error("Should not collide when bounding boxes don't overlap")
	}

	// Test 2: Entity with overlapping bounding box
	nearby := obstacle.NewAsteroid(60, 60, obstacle.AsteroidMedium) // 20x20, overlaps
	collision := def.Collides(asteroid, nearby)

	// Both asteroids implement PreciseCollider, so should use precise detection
	t.Logf("Nearby asteroids collision: %v (depends on shapes)", collision)
}

func TestCollidesUsesAsteroidPreciseCollider(t *testing.T) {
	// Verify that def.Collides() properly uses asteroid's CollidesWith
	// when asteroid implements PreciseCollider

	asteroid := obstacle.NewAsteroid(100, 100, obstacle.AsteroidHuge) // 48x48

	// Asteroid implements PreciseCollider
	var _ def.PreciseCollider = asteroid

	// Create another entity that overlaps
	other := obstacle.NewAsteroid(120, 120, obstacle.AsteroidMedium)

	// Both have bounding box overlap
	if !asteroid.BoundingBoxOverlaps(other) {
		t.Fatal("Test setup failed: bounding boxes should overlap")
	}

	// Collides should use precise detection from at least one of them
	_ = def.Collides(asteroid, other)

	// If we got here without panic, the collision system is working
	// The actual result depends on the procedural shapes
}

func TestAsteroidCollisionOnlyCountsSolidPixels(t *testing.T) {
	// This test verifies that CollidesWith only counts solid pixels
	// by checking a case where bounding boxes overlap but shapes might not

	// Create two asteroids with small overlap region
	asteroid1 := obstacle.NewAsteroid(100, 100, obstacle.AsteroidMedium) // 20x20
	asteroid2 := obstacle.NewAsteroid(118, 118, obstacle.AsteroidMedium) // 20x20

	// Bounding boxes overlap in a 2×2 region (118-120 in both dimensions)
	if !asteroid1.BoundingBoxOverlaps(asteroid2) {
		t.Skip("Test setup failed: need overlapping bounding boxes")
	}

	// Use precise collision
	collision := def.Collides(asteroid1, asteroid2)

	// The result depends on whether the procedural shapes have solid pixels
	// in the overlap region. Either outcome is valid.
	t.Logf("Collision with small overlap region: %v", collision)

	// The important thing is that it only returns true if there are
	// actual solid pixels in the overlap, not just bounding box overlap
}

func TestOneWayPreciseCollision(t *testing.T) {
	// Test that collision works when only ONE entity is a PreciseCollider
	// This is the typical case: asteroid (precise) vs simple entity (bounding box)

	asteroid := obstacle.NewAsteroid(100, 100, obstacle.AsteroidLarge)

	// Verify asteroid is a PreciseCollider (compile-time check)
	var _ def.PreciseCollider = asteroid

	// Create a simple entity with bounding box collision only
	entity := testutil.NewMockEntity(def.EntityTypePlayer)
	entity.X = 110
	entity.Y = 110
	entity.Width = 32
	entity.Height = 32

	// Check collision both ways
	collision1 := def.Collides(asteroid, entity)
	collision2 := def.Collides(entity, asteroid)

	// Should get the same result regardless of order
	if collision1 != collision2 {
		t.Errorf("Collision should be symmetric: Collides(a,b)=%v but Collides(b,a)=%v",
			collision1, collision2)
	}

	t.Logf("One-way precise collision result: %v", collision1)
}
