package player

import (
	"tsumegolang/pkg/tool/sprite"
)

type Hull struct {
	sprite *sprite.Sprite
	// HP contributed by this hull to the ship's max HP pool.
	// Multiple hulls or upgrade components are summed by the player
	// when computing its starting maxHP.
	HP int
}

// Basic

func BasicHull() (*Hull, error) {
	// Load basic hull
	hullData, err := spriteFiles.ReadFile("sprites/hull_basic.yaml")
	if err != nil {
		return nil, err
	}

	s, err := sprite.SpriteFromBytes(hullData)
	if err != nil {
		return nil, err
	}

	return &Hull{
		sprite: s,
		HP:     3,
	}, nil
}

// Add more hulls here as needed
