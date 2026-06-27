package effects

import (
	"testing"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/testutil"
)

func TestExplosionSmall(t *testing.T) {
	scene := testutil.NewMockScene()
	explosion, err := NewExplosion(100, 100, scene, def.ExplosionSmall)
	if err != nil {
		t.Fatalf("Failed to create small explosion: %v", err)
	}

	// Check dimensions (should be 16×16)
	width, height := explosion.Dimensions()
	if width != 16 || height != 16 {
		t.Errorf("Expected dimensions 16×16, got %d×%d", width, height)
	}

	// Check entity type
	if explosion.Type() != def.EntityTypeEnvironment {
		t.Errorf("Expected EntityTypeEnvironment, got %v", explosion.Type())
	}

	// Check it starts alive
	if explosion.CanBeRemoved() {
		t.Error("Explosion should not be removable immediately after creation")
	}

	// Simulate frames until expiration (40 frames for small)
	for i := 0; i < 39; i++ {
		explosion.Act(scene)
		if explosion.CanBeRemoved() {
			t.Errorf("Explosion became removable at frame %d, expected 40", i+1)
		}
	}

	// After 40 frames, should be removable
	explosion.Act(scene)
	if !explosion.CanBeRemoved() {
		t.Error("Explosion should be removable after 40 frames")
	}
}

func TestExplosionMedium(t *testing.T) {
	scene := testutil.NewMockScene()
	explosion, err := NewExplosion(200, 200, scene, def.ExplosionMedium)
	if err != nil {
		t.Fatalf("Failed to create medium explosion: %v", err)
	}

	// Check dimensions (should be 32×32)
	width, height := explosion.Dimensions()
	if width != 32 || height != 32 {
		t.Errorf("Expected dimensions 32×32, got %d×%d", width, height)
	}

	// Simulate frames until expiration (60 frames for medium)
	for i := 0; i < 59; i++ {
		explosion.Act(scene)
		if explosion.CanBeRemoved() {
			t.Errorf("Explosion became removable at frame %d, expected 60", i+1)
		}
	}

	// After 60 frames, should be removable
	explosion.Act(scene)
	if !explosion.CanBeRemoved() {
		t.Error("Explosion should be removable after 60 frames")
	}
}

func TestExplosionLarge(t *testing.T) {
	scene := testutil.NewMockScene()
	explosion, err := NewExplosion(300, 300, scene, def.ExplosionLarge)
	if err != nil {
		t.Fatalf("Failed to create large explosion: %v", err)
	}

	// Check dimensions (should be 48×48)
	width, height := explosion.Dimensions()
	if width != 48 || height != 48 {
		t.Errorf("Expected dimensions 48×48, got %d×%d", width, height)
	}

	// Simulate frames until expiration (96 frames for large)
	for i := 0; i < 95; i++ {
		explosion.Act(scene)
		if explosion.CanBeRemoved() {
			t.Errorf("Explosion became removable at frame %d, expected 96", i+1)
		}
	}

	// After 96 frames, should be removable
	explosion.Act(scene)
	if !explosion.CanBeRemoved() {
		t.Error("Explosion should be removable after 96 frames")
	}
}

func TestExplosionCentering(t *testing.T) {
	scene := testutil.NewMockScene()
	spawnX, spawnY := 100, 100

	explosion, err := NewExplosion(spawnX, spawnY, scene, def.ExplosionSmall)
	if err != nil {
		t.Fatalf("Failed to create explosion: %v", err)
	}

	// Explosion should be centered on spawn location
	// For 16×16 sprite, x should be 100 - 8 = 92, y should be 100 - 8 = 92
	x, y := explosion.Location()
	width, height := explosion.Dimensions()

	centerX := x + width/2
	centerY := y + height/2

	if centerX != spawnX {
		t.Errorf("Explosion not centered on X: expected center %d, got %d", spawnX, centerX)
	}
	if centerY != spawnY {
		t.Errorf("Explosion not centered on Y: expected center %d, got %d", spawnY, centerY)
	}
}

func TestExplosionNoCollision(t *testing.T) {
	scene := testutil.NewMockScene()
	explosion, err := NewExplosion(100, 100, scene, def.ExplosionSmall)
	if err != nil {
		t.Fatalf("Failed to create explosion: %v", err)
	}

	// Explosions should never collide with anything
	entity := testutil.NewMockEntity(def.EntityTypePlayer)
	if explosion.BoundingBoxOverlaps(entity) {
		t.Error("Explosion should not report bounding box overlaps")
	}
}
