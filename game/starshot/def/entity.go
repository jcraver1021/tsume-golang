package def

import (
	ebit "github.com/hajimehoshi/ebiten/v2"
)

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

// Top to bottom for update order
// Bottom to top for draw order
// (note that some entity types opt out of being drawn)
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
