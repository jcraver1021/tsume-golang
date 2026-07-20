package play

import (
	"math/rand"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/enemy"
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
	b.Entities().Add(environment.NewAsteroidField(0.01, func() obstacle.AsteroidSize {
		if rand.Float64() < 0.7 {
			return obstacle.AsteroidMassive
		}
		return obstacle.AsteroidColossal
	}))

	// Player entity - centered at bottom
	player, err := player.NewPlayer(def.ScreenWidth/2, def.ScreenHeight-50)
	if err != nil {
		panic(err)
	}
	b.Entities().Add(player)

	// Enemy chaser - starts near the top center
	chaser, err := enemy.NewChaser(def.ScreenWidth/2-7, 40)
	if err != nil {
		panic(err)
	}
	b.Entities().Add(chaser)
}
