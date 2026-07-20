package ui_test

import (
	"testing"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/ui"
	"tsumegolang/game/starshot/testutil"
)

func newTestHUD(wave, score int) *ui.HUD {
	return ui.NewHUD(&testutil.MockGameStateReader{Wave: wave, Score: score})
}

// --- Interface compliance ---

func TestHUDImplementsEntity(t *testing.T) {
	var _ def.Entity = newTestHUD(1, 0)
}

// --- Entity contract ---

func TestHUDTypeIsUI(t *testing.T) {
	if got := newTestHUD(1, 0).Type(); got != def.EntityTypeUI {
		t.Errorf("Type() = %v, want EntityTypeUI", got)
	}
}

func TestHUDCanNeverBeRemoved(t *testing.T) {
	if newTestHUD(1, 0).CanBeRemoved() {
		t.Error("CanBeRemoved() = true, HUD must persist for the entire scene lifetime")
	}
}

func TestHUDLocationIsOrigin(t *testing.T) {
	x, y := newTestHUD(1, 0).Location()
	if x != 0 || y != 0 {
		t.Errorf("Location() = (%d, %d), want (0, 0)", x, y)
	}
}

func TestHUDWidthMatchesScreen(t *testing.T) {
	w, _ := newTestHUD(1, 0).Dimensions()
	if w != def.ScreenWidth {
		t.Errorf("Dimensions() width = %d, want ScreenWidth %d", w, def.ScreenWidth)
	}
}

func TestHUDBoundingBoxNeverOverlaps(t *testing.T) {
	hud := newTestHUD(1, 0)
	other := testutil.NewMockEntity(def.EntityTypeEnemy)
	other.X, other.Y, other.Width, other.Height = 0, 0, def.ScreenWidth, def.ScreenHeight
	if hud.BoundingBoxOverlaps(other) {
		t.Error("BoundingBoxOverlaps() = true; HUD should never participate in collision")
	}
}

// --- Act: ammo reads ---

func TestHUDActWithNoPlayerDoesNotPanic(t *testing.T) {
	hud := newTestHUD(1, 100)
	scene := testutil.NewMockScene()
	hud.Act(scene) // no player entity — should be a no-op, not a panic
}

func TestHUDActWithPlayerThatLacksAmmoDoesNotPanic(t *testing.T) {
	hud := newTestHUD(1, 100)
	scene := testutil.NewMockScene()
	// Regular MockEntity implements Entity but not ammoReader
	scene.Entities().Add(testutil.NewMockEntity(def.EntityTypePlayer))
	hud.Act(scene)
}

func TestHUDActReadsSecondaryAmmoFromPlayer(t *testing.T) {
	hud := newTestHUD(2, 500)
	scene := testutil.NewMockScene()

	player := testutil.NewMockAmmoPlayer()
	player.CurrentAmmo = 3
	player.MaxAmmo = 10
	player.HasWeapon = true
	scene.Entities().Add(player)

	// Act should not panic when player satisfies ammoReader
	hud.Act(scene)
}

func TestHUDActWithExhaustedAmmoDoesNotPanic(t *testing.T) {
	hud := newTestHUD(1, 0)
	scene := testutil.NewMockScene()

	player := testutil.NewMockAmmoPlayer()
	player.CurrentAmmo = 0
	player.MaxAmmo = 10
	player.HasWeapon = true
	scene.Entities().Add(player)

	hud.Act(scene)
}

func TestHUDActWithNoSecondaryWeaponDoesNotPanic(t *testing.T) {
	hud := newTestHUD(1, 0)
	scene := testutil.NewMockScene()

	player := testutil.NewMockAmmoPlayer()
	player.HasWeapon = false
	scene.Entities().Add(player)

	hud.Act(scene)
}

// --- MockMortalEntity and MockDamageableEntity sanity checks ---
// These verify the mock contracts that collision and damage tests rely on.

func TestMockMortalStartsAlive(t *testing.T) {
	e := testutil.NewMockMortalEntity(def.EntityTypeEnemy)
	if e.IsDead() {
		t.Error("MockMortalEntity should start alive")
	}
}

func TestMockMortalDiesAfterMarkAsDead(t *testing.T) {
	e := testutil.NewMockMortalEntity(def.EntityTypeEnemy)
	e.MarkAsDead(testutil.NewMockScene())
	if !e.IsDead() {
		t.Error("MockMortalEntity should be dead after MarkAsDead")
	}
}

func TestMockDamageableStartsAtMaxHP(t *testing.T) {
	e := testutil.NewMockDamageableEntity(def.EntityTypeEnemy, 5)
	if e.CurrentHP() != e.MaxHP() {
		t.Errorf("CurrentHP %d != MaxHP %d at creation", e.CurrentHP(), e.MaxHP())
	}
}

func TestMockDamageableTakeDamageReducesHP(t *testing.T) {
	e := testutil.NewMockDamageableEntity(def.EntityTypeEnemy, 5)
	e.TakeDamage(2)
	if e.CurrentHP() != 3 {
		t.Errorf("HP after TakeDamage(2): got %d, want 3", e.CurrentHP())
	}
}

func TestMockDamageableHPFloorsAtZero(t *testing.T) {
	e := testutil.NewMockDamageableEntity(def.EntityTypeEnemy, 3)
	e.TakeDamage(9999)
	if e.CurrentHP() < 0 {
		t.Errorf("HP should not go negative, got %d", e.CurrentHP())
	}
}

func TestMockImpulsableRecordsLastImpulse(t *testing.T) {
	e := testutil.NewMockImpulsableEntity(def.EntityTypeObstacle)
	e.ApplyImpulse(5.0, -2.0)
	if e.LastDVX != 5.0 || e.LastDVY != -2.0 {
		t.Errorf("impulse recorded as (%.1f, %.1f), want (5.0, -2.0)", e.LastDVX, e.LastDVY)
	}
	if e.ImpulseCount != 1 {
		t.Errorf("ImpulseCount = %d, want 1", e.ImpulseCount)
	}
}
