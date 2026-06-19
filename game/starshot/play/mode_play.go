package play

import (
	"math/rand"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/environment"
	"tsumegolang/game/starshot/entity/obstacle"
	"tsumegolang/game/starshot/entity/player"
)

var (
	waveStarDensity = map[int]float64{
		1: 0.1,
	}
)

func initPlayMode(b def.Scene, state *GameState) {
	switch state.Wave {
	case 1:
		initWave1(b)
	}
}

func initWave1(b def.Scene) {
	// Starfield background - use wave-specific density
	density := waveStarDensity[1]
	b.Entities().Add(environment.NewSpace(density, b))

	// Player entity - centered at bottom
	player := player.NewPlayer(def.ScreenWidth/2, def.ScreenHeight-50)
	b.Entities().Add(player)

	// Spawn initial asteroids
	spawnAsteroids(b, 5)
}

// spawnAsteroids creates the specified number of asteroids at random positions
func spawnAsteroids(b def.Scene, count int) {
	for range count {
		// Random x position across screen width
		x := rand.Intn(def.ScreenWidth)
		// Start above screen
		y := -rand.Intn(def.ScreenHeight / 2)

		asteroid := obstacle.NewRandomAsteroid(x, y)
		b.Entities().Add(asteroid)
	}
}
