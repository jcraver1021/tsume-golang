package player

import (
	"tsumegolang/pkg/tool/sprite"
)

type EngineMount int

const (
	EngineMountCenter EngineMount = iota
)

type Engine struct {
	EngineMount EngineMount
	vUp         int
	vDown       int
	vLeft       int
	vRight      int
	sprite      *sprite.Sprite
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

	s, err := sprite.SpriteFromBytes(engineData)
	if err != nil {
		return nil, err
	}

	return &Engine{
		EngineMount: EngineMountCenter,
		vUp:         basicEngineSpeed,
		vDown:       basicEngineSpeed,
		vLeft:       basicEngineSpeed,
		vRight:      basicEngineSpeed,
		sprite:      s,
	}, nil
}

// Add more engines here as needed
