package def

import (
	ebit "github.com/hajimehoshi/ebiten/v2"
)

// EntityType categorizes entities for rendering and update order
type EntityType int

const (
	EntityTypeUI EntityType = iota // UI overlays (drawn on top of everything)
	EntityTypeEnvironment
	EntityTypePlayer
	EntityTypeTeam
	EntityTypeEnemy
	EntityTypeObstacle
	EntityTypeBackground
)

var EntityTypes = []EntityType{
	EntityTypeUI,
	EntityTypeEnvironment,
	EntityTypePlayer,
	EntityTypeTeam,
	EntityTypeEnemy,
	EntityTypeObstacle,
	EntityTypeBackground,
}

var EntityTypeNames = map[EntityType]string{
	EntityTypeUI:          "UI",
	EntityTypeEnvironment: "Environment",
	EntityTypePlayer:      "Player",
	EntityTypeTeam:        "Team",
	EntityTypeEnemy:       "Enemy",
	EntityTypeObstacle:    "Obstacle",
	EntityTypeBackground:  "Background",
}

// Screen constants - shared by all entities
const (
	ScreenWidth  = 480
	ScreenHeight = 640
)

// OnScreen indicates an object's visibility relative to the screen boundaries
type OnScreen int

const (
	Fully OnScreen = iota
	Partially
	OffScreen
)

// Scene provides the game state context for entity actions
type Scene interface {
	Width() int
	Height() int
	Entities() EntityCollection
	Tick() int // Global animation tick counter
}

// Entity is the core interface that all game objects must implement
type Entity interface {
	Type() EntityType
	Location() (x, y int)
	Dimensions() (width, height int)
	// BoundingBoxOverlaps performs fast AABB collision check (broad phase)
	// Returns true if bounding boxes might be touching
	BoundingBoxOverlaps(other Entity) bool
	Act(Scene)
	Draw(*ebit.Image)
	CanBeRemoved() bool
}

// PreciseCollider is an optional interface for entities that need
// pixel-perfect or shape-based collision detection (narrow phase)
type PreciseCollider interface {
	Entity
	// CollidesWith performs precise collision detection
	// Only called after BoundingBoxOverlaps returns true
	CollidesWith(other Entity) bool
}

// ExplosionSize specifies the visual scale of an explosion effect
type ExplosionSize int

const (
	ExplosionSmall ExplosionSize = iota
	ExplosionMedium
	ExplosionLarge
)

// DeathEffect specifies what happens when an entity dies
type DeathEffect struct {
	ExplosionSize      ExplosionSize
	SlowdownMultiplier float64 // 0.0 = no slowdown, 0.3 = 30% speed, 1.0 = normal
	SlowdownDuration   int     // Frames to maintain slowdown (0 = no slowdown)
}

// Mortal is an optional interface for entities that can die and spawn effects
type Mortal interface {
	Entity
	GetDeathEffect() DeathEffect
	MarkAsDead(scene Scene)
	IsDead() bool
}

// EntityCollection provides access to entities without exposing implementation details
type EntityCollection interface {
	Add(Entity)
	Get(EntityType) []Entity
	IterateForUpdate() <-chan Entity
	IterateForDraw() <-chan Entity
}

// Damageable is an optional interface for entities with hit points.
// TakeDamage reduces HP; callers should then check IsDead() via Mortal.
type Damageable interface {
	Entity
	TakeDamage(amount int)
	CurrentHP() int
	MaxHP() int
}

// Impulsable is an optional interface for entities that can receive
// a velocity impulse from projectile hits.
type Impulsable interface {
	Entity
	ApplyImpulse(dvx, dvy float64)
}

// Collides performs two-phase collision detection between entities
// Phase 1: Fast bounding box check via BoundingBoxOverlaps()
// Phase 2: Precise check via CollidesWith() if at least one implements PreciseCollider
// Returns true only if entities are actually colliding
func Collides(a, b Entity) bool {
	// Broad phase: cheap bounding box check
	if !a.BoundingBoxOverlaps(b) {
		return false
	}

	// Narrow phase: if either entity has precise collision, use it
	preciseA, aHasPrecise := a.(PreciseCollider)
	preciseB, bHasPrecise := b.(PreciseCollider)

	if aHasPrecise {
		// A has precise collision - use it
		return preciseA.CollidesWith(b)
	}

	if bHasPrecise {
		// B has precise collision - use it
		return preciseB.CollidesWith(a)
	}

	// Neither has precise collision - bounding box overlap is sufficient
	return true
}
