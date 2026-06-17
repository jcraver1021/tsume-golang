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
}

// Entity is the core interface that all game objects must implement
type Entity interface {
	Type() EntityType
	Location() (x, y int)
	Dimensions() (width, height int)
	Overlaps(other Entity) bool
	Act(Scene)
	Draw(*ebit.Image)
	CanBeRemoved() bool
}

// EntityCollection provides access to entities without exposing implementation details
type EntityCollection interface {
	Add(Entity)
	Get(EntityType) []Entity
	IterateForUpdate() <-chan Entity
	IterateForDraw() <-chan Entity
}
