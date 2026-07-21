package play

import (
	"math/rand"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/enemy"
	"tsumegolang/game/starshot/entity/environment"
	"tsumegolang/game/starshot/entity/obstacle"
	"tsumegolang/game/starshot/entity/player"
	"tsumegolang/game/starshot/entity/ui"
)

func initPlayMode(b def.Scene, state *GameState) {
	state.Score = 0

	b.Entities().Add(environment.NewSpace(0.1, b))
	b.Entities().Add(environment.NewAsteroidField(0.01, func() obstacle.AsteroidSize {
		if rand.Float64() < 0.7 {
			return obstacle.AsteroidMassive
		}
		return obstacle.AsteroidColossal
	}))

	gun, err := player.NewBasicGun()
	if err != nil {
		panic(err)
	}
	launcher, err := player.NewBombLauncher()
	if err != nil {
		panic(err)
	}
	p, err := player.NewPlayer(def.ScreenWidth/2, def.ScreenHeight-50, gun, launcher)
	if err != nil {
		panic(err)
	}
	b.Entities().Add(p)
	b.Entities().Add(ui.NewHUD(state))

	spawnWaveEnemies(b, state)
}

// spawnWaveEnemies adds the enemy entities for the current wave number.
// Called on wave start and whenever the player clears a wave mid-game.
func spawnWaveEnemies(b def.Scene, state *GameState) {
	switch state.Wave {
	case 1:
		spawnWave1(b)
	default:
		spawnWave2(b)
	}
}

// spawnWave1 spawns three Chasers plus a spread of drifters and weavers.
func spawnWave1(b def.Scene) {
	xs := []int{def.ScreenWidth / 4, def.ScreenWidth / 2, 3 * def.ScreenWidth / 4}
	for _, x := range xs {
		c, err := enemy.NewChaser(x, 40)
		if err != nil {
			panic(err)
		}
		b.Entities().Add(c)
	}

	// Drifters: fast straight-down threats on the flanks
	for _, x := range []int{60, 180, 300, 420} {
		d, err := enemy.NewDrifter(x, rand.Intn(30)+10)
		if err != nil {
			panic(err)
		}
		b.Entities().Add(d)
	}

	// Weavers: navigate the asteroid field while descending
	for _, x := range []int{120, 360} {
		wv, err := enemy.NewWeaver(x, rand.Intn(30)+10)
		if err != nil {
			panic(err)
		}
		b.Entities().Add(wv)
	}
}

// spawnWave2 spawns two Hunters plus one of each mine type.
func spawnWave2(b def.Scene) {
	for _, x := range []int{def.ScreenWidth / 3, 2 * def.ScreenWidth / 3} {
		h, err := enemy.NewHunter(x, 40)
		if err != nil {
			panic(err)
		}
		b.Entities().Add(h)
	}

	// Contact mine — chases and detonates on touch
	cm, err := enemy.NewMine(80, rand.Intn(60)+30)
	if err != nil {
		panic(err)
	}
	b.Entities().Add(cm)

	// Range mine — detonates after player lingers within blast radius for 10 frames
	rm, err := enemy.NewRangeMine(def.ScreenWidth/2, rand.Intn(60)+30)
	if err != nil {
		panic(err)
	}
	b.Entities().Add(rm)

	// Path mine — zigzag pattern, blue light, detonates on contact
	zigzag := []enemy.PathSegment{
		{Frames: 80, VX: -1.2, VY: 0},
		{Frames: 80, VX: 1.2, VY: 0},
	}
	pm, err := enemy.NewPathMine(400, rand.Intn(60)+30, zigzag)
	if err != nil {
		panic(err)
	}
	b.Entities().Add(pm)

	// Path range mine — sweeps side-to-side, violet lights, detonates after player lingers 10 s
	sweep := []enemy.PathSegment{
		{Frames: 120, VX: -0.8, VY: 0},
		{Frames: 120, VX: 0.8, VY: 0},
	}
	prm, err := enemy.NewPathRangeMine(def.ScreenWidth/2, rand.Intn(40)+60, sweep)
	if err != nil {
		panic(err)
	}
	b.Entities().Add(prm)
}
