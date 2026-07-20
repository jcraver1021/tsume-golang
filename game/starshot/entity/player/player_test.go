package player_test

import (
	"testing"

	"tsumegolang/game/starshot/entity/player"
)

// newTestPlayer creates a player with no weapon (nil is handled gracefully).
func newTestPlayer(t *testing.T) *player.Player {
	t.Helper()
	p, err := player.NewPlayer(240, 400, nil, nil)
	if err != nil {
		t.Fatalf("NewPlayer returned error: %v", err)
	}
	return p
}

// --- Hull HP contribution ---

func TestPlayerHPDerivedFromHull(t *testing.T) {
	p := newTestPlayer(t)
	// BasicHull contributes 3 HP; player should start at that value.
	if p.CurrentHP() <= 0 {
		t.Errorf("player should start with positive HP, got %d", p.CurrentHP())
	}
	if p.CurrentHP() != p.MaxHP() {
		t.Errorf("CurrentHP %d should equal MaxHP %d at creation", p.CurrentHP(), p.MaxHP())
	}
}

func TestPlayerMaxHPMatchesBasicHull(t *testing.T) {
	p := newTestPlayer(t)
	const basicHullHP = 3
	if p.MaxHP() != basicHullHP {
		t.Errorf("MaxHP = %d, want %d (BasicHull HP)", p.MaxHP(), basicHullHP)
	}
}

// --- AddMaxHP ---

func TestAddMaxHPIncreasesMaxHP(t *testing.T) {
	p := newTestPlayer(t)
	before := p.MaxHP()
	p.AddMaxHP(2)
	if p.MaxHP() != before+2 {
		t.Errorf("MaxHP after AddMaxHP(2): got %d, want %d", p.MaxHP(), before+2)
	}
}

func TestAddMaxHPAlsoIncreasesCurrentHP(t *testing.T) {
	p := newTestPlayer(t)
	before := p.CurrentHP()
	p.AddMaxHP(5)
	if p.CurrentHP() != before+5 {
		t.Errorf("CurrentHP after AddMaxHP(5): got %d, want %d", p.CurrentHP(), before+5)
	}
}

func TestAddMaxHPPreservesExistingDamage(t *testing.T) {
	p := newTestPlayer(t)
	p.TakeDamage(1)
	damagedHP := p.CurrentHP()
	p.AddMaxHP(3)
	// Current HP should rise by 3 (bonus), not reset to new max
	if p.CurrentHP() != damagedHP+3 {
		t.Errorf("CurrentHP after damage+upgrade: got %d, want %d", p.CurrentHP(), damagedHP+3)
	}
}

// --- TakeDamage ---

func TestPlayerTakeDamageReducesHP(t *testing.T) {
	p := newTestPlayer(t)
	before := p.CurrentHP()
	p.TakeDamage(1)
	if p.CurrentHP() != before-1 {
		t.Errorf("HP after TakeDamage(1): got %d, want %d", p.CurrentHP(), before-1)
	}
}

func TestPlayerDiesAtZeroHP(t *testing.T) {
	p := newTestPlayer(t)
	p.TakeDamage(p.MaxHP())
	if !p.IsDead() {
		t.Error("player should be dead after lethal damage")
	}
	if p.CurrentHP() != 0 {
		t.Errorf("HP after death should be 0, got %d", p.CurrentHP())
	}
}

func TestPlayerTakeDamageIgnoredWhenDead(t *testing.T) {
	p := newTestPlayer(t)
	p.TakeDamage(p.MaxHP()) // kill
	p.TakeDamage(1)         // should be no-op
	if p.CurrentHP() != 0 {
		t.Errorf("dead player HP should stay 0, got %d", p.CurrentHP())
	}
}
