package player

import (
	"tsumegolang/game/starshot/draw"
)

type Hull struct {
	sprite *draw.ColorMatrix
}

// Basic

func BasicHull() (*Hull, error) {
	// Load basic hull
	hullData, err := spriteFiles.ReadFile("sprites/hull_basic.yaml")
	if err != nil {
		return nil, err
	}

	sprite, err := draw.ColorMatrixFromBytes(hullData)
	if err != nil {
		return nil, err
	}

	return &Hull{
		sprite: sprite,
	}, nil
}

// Add more hulls here as needed
