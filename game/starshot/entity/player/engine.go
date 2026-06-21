package player

import (
	"tsumegolang/game/starshot/draw"
)


type EngineMount int

const (
	EngineMountCenter EngineMount = iota
)

type Engine struct {
	EngineMount EngineMount
	vUp int
	vDown int
	vLeft int
	vRight int
	sprite *draw.ColorMatrix
}

// Basic

const (
	basicEngineSpeed = 5
)

func BasicEngine() (*Engine, error) {
	// Load basic engine
	engineData, err := spriteFiles.ReadFile("sprites/engine_basic.yaml")
	if err != nil {
		return nil, err
	}

	sprite, err := draw.ColorMatrixFromBytes(engineData)
	if err != nil {
		return nil, err
	}

	return &Engine{
		EngineMount: EngineMountCenter,
		vUp:         basicEngineSpeed,
		vDown:       basicEngineSpeed,
		vLeft:       basicEngineSpeed,
		vRight:      basicEngineSpeed,
		sprite:      sprite,
	}, nil
}

// Add more engines here as needed