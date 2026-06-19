package def

import (
	ebit "github.com/hajimehoshi/ebiten/v2"
)

// EntityType categorizes entities for rendering and update order
type EntityType int

const (
	EntityTypeEnvironment EntityType = iota
	EntityTypePlayer
	EntityTypeTeam
	EntityTypeEnemy
	EntityTypeObstacle
	EntityTypeBackground
)

var EntityTypes = []EntityType{
	EntityTypeEnvironment,
	EntityTypePlayer,
	EntityTypeTeam,
	EntityTypeEnemy,
	EntityTypeObstacle,
	EntityTypeBackground,
}

var EntityTypeNames = map[EntityType]string{
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

// OnScreen indicates whether an entity is visible on screen
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

// EntityCollection provides access to entities without exposing implementation details
type EntityCollection interface {
	Add(Entity)
	Get(EntityType) []Entity
	IterateForUpdate() <-chan Entity
	IterateForDraw() <-chan Entity
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
