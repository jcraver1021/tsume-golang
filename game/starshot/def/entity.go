package def

import (
	ebit "github.com/hajimehoshi/ebiten/v2"
)

// EntityType categorizes entities for update and draw order.
// Iota values are arbitrary — draw/update order is defined by EntityTypes below.
type EntityType int

const (
	EntityTypeUI EntityType = iota
	EntityTypeWave
	EntityTypeEnvironment
	EntityTypePlayer
	EntityTypeTeam // player projectiles
	EntityTypeEnemy
	EntityTypeEnemyTeam // enemy projectiles
	EntityTypeObstacle
	EntityTypeBackground
)

// EntityTypes defines update order (index 0 first) and, reversed, draw order
// (index 0 on top). UI runs first in update and renders last (above everything).
// Background runs last in update and renders first (behind everything).
var EntityTypes = []EntityType{
	EntityTypeUI,
	EntityTypeWave,
	EntityTypeEnvironment,
	EntityTypePlayer,
	EntityTypeTeam,
	EntityTypeEnemy,
	EntityTypeEnemyTeam,
	EntityTypeObstacle,
	EntityTypeBackground,
}

var EntityTypeNames = map[EntityType]string{
	EntityTypeUI:          "UI",
	EntityTypeWave:        "Wave",
	EntityTypeEnvironment: "Environment",
	EntityTypePlayer:      "Player",
	EntityTypeTeam:        "Team",
	EntityTypeEnemy:       "Enemy",
	EntityTypeEnemyTeam:   "EnemyTeam",
	EntityTypeObstacle:    "Obstacle",
	EntityTypeBackground:  "Background",
}

// Entity is the core interface that all game objects must implement
type Entity interface {
	Type() EntityType
	Location() (x, y int)
	Dimensions() (width, height int)
	// BoundingBoxOverlaps performs fast AABB collision check (broad phase).
	// Returns true if bounding boxes overlap; used to gate precise checks.
	BoundingBoxOverlaps(other Entity) bool
	Act(Scene)
	Draw(*ebit.Image)
	CanBeRemoved() bool
}
