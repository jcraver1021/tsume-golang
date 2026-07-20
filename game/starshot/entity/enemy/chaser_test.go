package enemy_test

import (
	"math"
	"testing"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/enemy"
	"tsumegolang/game/starshot/entity/obstacle"
	"tsumegolang/game/starshot/testutil"
)

// newTestChaser creates a chaser at the given position, failing the test on error.
func newTestChaser(t *testing.T, x, y int) *enemy.Chaser {
	t.Helper()
	c, err := enemy.NewChaser(x, y)
	if err != nil {
		t.Fatalf("NewChaser returned error: %v", err)
	}
	return c
}

// distance returns the Euclidean distance between two points.
func distance(x1, y1, x2, y2 int) float64 {
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	return math.Sqrt(dx*dx + dy*dy)
}

// centerOf returns the center of an entity's bounding box.
func centerOf(e def.Entity) (cx, cy int) {
	x, y := e.Location()
	w, h := e.Dimensions()
	return x + w/2, y + h/2
}

// --- Movement toward player ---

func TestChaserMovesTowardPlayer(t *testing.T) {
	// Chaser starts at top-center; player is at bottom-center.
	const chaserX, chaserY = 233, 50
	const playerX, playerY = 233, 560

	scene := testutil.NewMockScene()
	player := testutil.NewMockEntity(def.EntityTypePlayer)
	player.X, player.Y, player.Width, player.Height = playerX, playerY, 20, 20
	scene.Entities().Add(player)

	chaser := newTestChaser(t, chaserX, chaserY)
	cx0, cy0 := centerOf(chaser)
	px, py := centerOf(player)
	dist0 := distance(cx0, cy0, px, py)

	for range 20 {
		chaser.Act(scene)
	}

	cx1, cy1 := centerOf(chaser)
	dist1 := distance(cx1, cy1, px, py)

	if dist1 >= dist0 {
		t.Errorf("chaser should have moved closer to player: dist before=%.1f, after=%.1f", dist0, dist1)
	}
}

func TestChaserMovesDownWhenPlayerIsBelow(t *testing.T) {
	scene := testutil.NewMockScene()
	player := testutil.NewMockEntity(def.EntityTypePlayer)
	player.X, player.Y, player.Width, player.Height = 233, 560, 20, 20
	scene.Entities().Add(player)

	chaser := newTestChaser(t, 233, 50)
	_, y0 := chaser.Location()

	chaser.Act(scene)

	_, y1 := chaser.Location()
	if y1 <= y0 {
		t.Errorf("chaser should move down toward player below: y0=%d y1=%d", y0, y1)
	}
}

func TestChaserDoesNotMoveWithoutPlayer(t *testing.T) {
	scene := testutil.NewMockScene() // no player added
	chaser := newTestChaser(t, 100, 100)
	x0, y0 := chaser.Location()

	chaser.Act(scene)

	x1, y1 := chaser.Location()
	if x0 != x1 || y0 != y1 {
		t.Errorf("chaser should not move without a player in the scene: (%d,%d) → (%d,%d)", x0, y0, x1, y1)
	}
}

// --- Dodge behavior ---

func TestChaserDodgesObstacleInPath(t *testing.T) {
	// Chaser at top-center, player directly below. A large obstacle is placed
	// between them within the lookahead distance so the dodge logic fires.
	//
	// Chaser sprite is 14×16, so center at (chaserX+7, chaserY+8).
	const chaserX, chaserY = 233, 100

	scene := testutil.NewMockScene()

	player := testutil.NewMockEntity(def.EntityTypePlayer)
	player.X, player.Y, player.Width, player.Height = 233, 560, 20, 20
	scene.Entities().Add(player)

	// A 48×48 obstacle placed directly ahead of the chaser.
	// Along the (≈0, 1) movement direction, "ahead" means higher Y.
	// Obstacle center at (240, 155) → 47px ahead of chaser center (240, 108).
	// Perp distance = 0 (directly in line), well within danger radius.
	obs := obstacle.NewAsteroid(216, 131, obstacle.AsteroidHuge) // 48×48, center ≈ (240, 155)
	scene.Entities().Add(obs)

	chaser := newTestChaser(t, chaserX, chaserY)
	x0, _ := chaser.Location()

	for range 5 {
		chaser.Act(scene)
	}

	x1, _ := chaser.Location()
	if x1 == x0 {
		t.Errorf("chaser should have moved laterally to dodge obstacle, but x stayed at %d", x0)
	}
}

// --- Damage and death ---

func TestChaserTakeDamageReducesHP(t *testing.T) {
	chaser := newTestChaser(t, 0, 0)
	before := chaser.CurrentHP()
	chaser.TakeDamage(1)
	if chaser.CurrentHP() != before-1 {
		t.Errorf("HP after TakeDamage(1): got %d, want %d", chaser.CurrentHP(), before-1)
	}
}

func TestChaserDiesAtZeroHP(t *testing.T) {
	chaser := newTestChaser(t, 0, 0)
	chaser.TakeDamage(chaser.MaxHP())
	if !chaser.IsDead() {
		t.Error("chaser should be dead after lethal damage")
	}
	if chaser.CurrentHP() != 0 {
		t.Errorf("HP should be 0 after death, got %d", chaser.CurrentHP())
	}
}

func TestChaserTakeDamageIgnoredWhenDead(t *testing.T) {
	chaser := newTestChaser(t, 0, 0)
	chaser.TakeDamage(chaser.MaxHP()) // kill
	chaser.TakeDamage(1)              // should be no-op
	if chaser.CurrentHP() != 0 {
		t.Errorf("dead chaser HP should stay 0, got %d", chaser.CurrentHP())
	}
}

func TestChaserOverkillDoesNotGoNegative(t *testing.T) {
	chaser := newTestChaser(t, 0, 0)
	chaser.TakeDamage(9999)
	if chaser.CurrentHP() < 0 {
		t.Errorf("HP should not go negative, got %d", chaser.CurrentHP())
	}
}

func TestChaserStopsMovingWhenDead(t *testing.T) {
	scene := testutil.NewMockScene()
	player := testutil.NewMockEntity(def.EntityTypePlayer)
	player.X, player.Y, player.Width, player.Height = 233, 560, 20, 20
	scene.Entities().Add(player)

	chaser := newTestChaser(t, 233, 50)
	chaser.TakeDamage(chaser.MaxHP()) // kill before moving

	x0, y0 := chaser.Location()
	chaser.Act(scene)
	x1, y1 := chaser.Location()

	if x0 != x1 || y0 != y1 {
		t.Errorf("dead chaser should not move: (%d,%d) → (%d,%d)", x0, y0, x1, y1)
	}
}

// --- Interface compliance ---

func TestChaserImplementsMortal(t *testing.T) {
	chaser := newTestChaser(t, 0, 0)
	var _ def.Mortal = chaser
}

func TestChaserImplementsDamageable(t *testing.T) {
	chaser := newTestChaser(t, 0, 0)
	var _ def.Damageable = chaser
}

// stubDrawTarget satisfies *ebit.Image for the Draw interface check without rendering.
// We never call Draw in these tests; this just ensures the type compiles.
var _ def.Entity = (*enemy.Chaser)(nil)
var _ = (*ebit.Image)(nil) // keep the import used
