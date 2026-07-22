package obstacle_test

import (
	"testing"

	"tsumegolang/game/starshot/entity/obstacle"
	"tsumegolang/game/starshot/testutil"
)

// --- HP values ---

func TestAsteroidSizeHPValues(t *testing.T) {
	testCases := []struct {
		size obstacle.AsteroidSize
		want int
	}{
		{obstacle.AsteroidTiny, 1},
		{obstacle.AsteroidSmall, 2},
		{obstacle.AsteroidMedium, 3},
		{obstacle.AsteroidLarge, 5},
		{obstacle.AsteroidHuge, 8},
		{obstacle.AsteroidMassive, 12},
		{obstacle.AsteroidGigantic, 18},
		{obstacle.AsteroidColossal, 25},
	}
	for _, tc := range testCases {
		if got := tc.size.HP(); got != tc.want {
			t.Errorf("size %d HP() = %d, want %d", tc.size, got, tc.want)
		}
	}
}

func TestAsteroidSizeMassStrictlyIncreases(t *testing.T) {
	prev := obstacle.AsteroidTiny.Mass()
	for s := obstacle.AsteroidSmall; s <= obstacle.AsteroidColossal; s++ {
		m := s.Mass()
		if m <= prev {
			t.Errorf("size %d mass %.0f should be greater than previous size's %.0f", s, m, prev)
		}
		prev = m
	}
}

// --- TakeDamage ---

func TestTakeDamageReducesHP(t *testing.T) {
	a := obstacle.NewAsteroid(0, 0, obstacle.AsteroidMedium) // 3 HP
	a.TakeDamage(1)
	if got := a.CurrentHP(); got != 2 {
		t.Errorf("after 1 damage: HP = %d, want 2", got)
	}
	if a.IsDead() {
		t.Error("should not be dead at 2 HP")
	}
}

func TestTakeDamageKillsAtZeroHP(t *testing.T) {
	a := obstacle.NewAsteroid(0, 0, obstacle.AsteroidTiny) // 1 HP
	a.TakeDamage(1)
	if !a.IsDead() {
		t.Error("should be dead after 1 damage on 1-HP asteroid")
	}
	if a.CurrentHP() != 0 {
		t.Errorf("HP should be 0 after death, got %d", a.CurrentHP())
	}
}

func TestTakeDamageOverkillDoesNotGoNegative(t *testing.T) {
	a := obstacle.NewAsteroid(0, 0, obstacle.AsteroidSmall) // 2 HP
	a.TakeDamage(1000)
	if a.CurrentHP() < 0 {
		t.Errorf("HP should not go negative, got %d", a.CurrentHP())
	}
}

func TestTakeDamageIgnoredWhenAlreadyDead(t *testing.T) {
	a := obstacle.NewAsteroid(0, 0, obstacle.AsteroidTiny)
	a.TakeDamage(1) // kill
	a.TakeDamage(1) // should be a no-op
	if a.CurrentHP() != 0 {
		t.Errorf("HP should stay 0 after second hit on dead asteroid, got %d", a.CurrentHP())
	}
}

func TestMaxHPMatchesSizeHPAtCreation(t *testing.T) {
	for _, size := range []obstacle.AsteroidSize{
		obstacle.AsteroidTiny, obstacle.AsteroidMedium, obstacle.AsteroidColossal,
	} {
		a := obstacle.NewAsteroid(0, 0, size)
		if a.CurrentHP() != a.MaxHP() {
			t.Errorf("size %d: CurrentHP %d != MaxHP %d at creation", size, a.CurrentHP(), a.MaxHP())
		}
		if a.MaxHP() != size.HP() {
			t.Errorf("size %d: MaxHP %d != size.HP() %d", size, a.MaxHP(), size.HP())
		}
	}
}

// --- Splitting ---

func TestSplitSpawnsTwoChildren(t *testing.T) {
	for _, size := range []obstacle.AsteroidSize{
		obstacle.AsteroidSmall,
		obstacle.AsteroidMedium,
		obstacle.AsteroidColossal,
	} {
		scene := testutil.NewMockScene()
		a := obstacle.NewAsteroid(100, 100, size)
		a.MarkAsDead(scene)
		entities := scene.GetEntities()
		if len(entities) != 2 {
			t.Errorf("size %d: expected 2 children after death, got %d", size, len(entities))
		}
	}
}

func TestTinyDoesNotSplit(t *testing.T) {
	scene := testutil.NewMockScene()
	a := obstacle.NewAsteroid(0, 0, obstacle.AsteroidTiny)
	a.MarkAsDead(scene)
	if n := len(scene.GetEntities()); n != 0 {
		t.Errorf("Tiny asteroid should not split, got %d children", n)
	}
}

func TestSplitChildrenAreSmallerSize(t *testing.T) {
	// Medium (20×20) must split into Small (12×12)
	scene := testutil.NewMockScene()
	a := obstacle.NewAsteroid(100, 100, obstacle.AsteroidMedium)
	a.MarkAsDead(scene)

	wantW, wantH := obstacle.AsteroidSmall.Dimensions()
	for i, child := range scene.GetEntities() {
		w, h := child.Dimensions()
		if w != wantW || h != wantH {
			t.Errorf("child %d: dimensions %dx%d, want %dx%d (Small)", i, w, h, wantW, wantH)
		}
	}
}

func TestSplitChildrenHaveDifferentXPositions(t *testing.T) {
	scene := testutil.NewMockScene()
	a := obstacle.NewAsteroid(100, 100, obstacle.AsteroidMedium)
	a.MarkAsDead(scene)

	entities := scene.GetEntities()
	if len(entities) != 2 {
		t.Fatalf("expected 2 children, got %d", len(entities))
	}
	x0, _ := entities[0].Location()
	x1, _ := entities[1].Location()
	if x0 == x1 {
		t.Error("split children should have different x positions")
	}
}

func TestSplitChildrenNearParentPosition(t *testing.T) {
	const parentX, parentY = 200, 300
	const parentW = 32 // AsteroidLarge
	scene := testutil.NewMockScene()
	a := obstacle.NewAsteroid(parentX, parentY, obstacle.AsteroidLarge)
	a.MarkAsDead(scene)

	parentCX := parentX + parentW/2
	for i, child := range scene.GetEntities() {
		cx, _ := child.Location()
		cw, _ := child.Dimensions()
		childCX := cx + cw/2
		// Children should start within a reasonable distance of the parent center
		if abs(childCX-parentCX) > 50 {
			t.Errorf("child %d center x %d too far from parent center x %d", i, childCX, parentCX)
		}
	}
}

// --- Impulse ---

func TestApplyImpulseCausesLateralMovement(t *testing.T) {
	a := obstacle.NewAsteroid(100, 100, obstacle.AsteroidTiny) // mass = 1
	x0, _ := a.Location()

	a.ApplyImpulse(8.0, 0)
	a.Act(testutil.NewMockScene())

	x1, _ := a.Location()
	if x1 <= x0 {
		t.Errorf("rightward impulse should increase x: x0=%d x1=%d", x0, x1)
	}
}

func TestApplyImpulseNegativePushesLeft(t *testing.T) {
	a := obstacle.NewAsteroid(100, 100, obstacle.AsteroidTiny)
	x0, _ := a.Location()

	a.ApplyImpulse(-8.0, 0)
	a.Act(testutil.NewMockScene())

	x1, _ := a.Location()
	if x1 >= x0 {
		t.Errorf("leftward impulse should decrease x: x0=%d x1=%d", x0, x1)
	}
}

func TestApplyImpulseSmallMovesFurtherThanLarge(t *testing.T) {
	const impulse = 64.0

	tiny := obstacle.NewAsteroid(100, 100, obstacle.AsteroidTiny)
	huge := obstacle.NewAsteroid(100, 100, obstacle.AsteroidHuge)

	tiny.ApplyImpulse(impulse, 0)
	huge.ApplyImpulse(impulse, 0)

	tiny.Act(testutil.NewMockScene())
	huge.Act(testutil.NewMockScene())

	tx, _ := tiny.Location()
	hx, _ := huge.Location()

	if tx <= hx {
		t.Errorf("tiny (x=%d) should have moved further right than huge (x=%d) from same impulse", tx, hx)
	}
}

func TestApplyImpulseVXCapEnforced(t *testing.T) {
	a := obstacle.NewAsteroid(100, 100, obstacle.AsteroidTiny)
	a.ApplyImpulse(100000.0, 0) // enormous push

	x0, _ := a.Location()
	a.Act(testutil.NewMockScene())
	x1, _ := a.Location()

	moved := x1 - x0
	if moved > 8 {
		t.Errorf("vx cap not enforced: moved %d px in one frame, max should be 8", moved)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
