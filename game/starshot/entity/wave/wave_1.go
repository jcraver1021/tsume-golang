package wave

import (
	"math/rand"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/enemy"
	"tsumegolang/game/starshot/entity/environment"
	"tsumegolang/game/starshot/entity/obstacle"
)

type Wave1 struct {
	ticksInPhase         int
	phase                int
	densityEarly         float64
	asteroidSizeEarly    func() obstacle.AsteroidSize
	enemies1             func() def.Entity
	enemies2             func() def.Entity
	minePathEarly        []enemy.PathSegment
	mineSpacingEarly     int
	minefieldWidthEarly  int
	minefieldHeightEarly int
	minefieldDrift       float64
}

func NewWave1() *Wave1 {
	return &Wave1{
		ticksInPhase: 0,
		phase:        0,
		densityEarly: 0.02,
		asteroidSizeEarly: func() obstacle.AsteroidSize {
			if rand.Float64() < 0.7 {
				return obstacle.AsteroidLarge
			}
			return obstacle.AsteroidHuge
		},
		enemies1: func() def.Entity {
			e, _ := enemy.NewDrifter(rand.Intn(def.ScreenWidth), -30)
			return e
		},
		enemies2: func() def.Entity {
			e, _ := enemy.NewWeaver(rand.Intn(def.ScreenWidth), -30)
			return e
		},
		minePathEarly: []enemy.PathSegment{
			{Frames: 60, VX: 0.0, VY: -1.0},
			{Frames: 90, VX: 1.0, VY: 0.0},
			{Frames: 60, VX: 0.0, VY: 1.0},
			{Frames: 90, VX: -1.0, VY: 0.0},
		},
		mineSpacingEarly:     120,
		minefieldWidthEarly:  4,
		minefieldHeightEarly: 4,
		minefieldDrift:       0,
	}
}

func (w *Wave1) Type() def.EntityType {
	return def.EntityTypeWave
}

func (w *Wave1) Onscreen() def.OnScreen {
	return def.OffScreen
}

func (w *Wave1) Location() (x, y int) {
	return 0, 0
}

func (w *Wave1) Dimensions() (width, height int) {
	return def.ScreenWidth, def.ScreenHeight
}

func (w *Wave1) BoundingBoxOverlaps(other def.Entity) bool {
	return false
}

func (w *Wave1) Act(b def.Scene) {
	switch w.phase {
	case 0:
		// Begin the first wave with some initial asteroids
		b.Entities().Add(environment.NewSpace(0.1, b))
		b.Entities().Add(obstacle.NewAsteroid(rand.Intn(def.ScreenWidth), -30, obstacle.AsteroidLarge))
		b.Entities().Add(obstacle.NewAsteroid(rand.Intn(def.ScreenWidth), -50, obstacle.AsteroidMedium))
		b.Entities().Add(obstacle.NewAsteroid(rand.Intn(def.ScreenWidth), -10, obstacle.AsteroidHuge))
		w.phase = 1
		w.resetTicks()
	case 1:
		switch w.ticksInPhase {
		case 300:
			// Begin the asteroid field after 5 seconds
			earlyAsteroidField := environment.NewAsteroidField(w.densityEarly, w.asteroidSizeEarly)
			b.Entities().Add(earlyAsteroidField)

			enemies1 := environment.NewSpawner(w.densityEarly, func() def.Entity {
				e, _ := enemy.NewDrifter(rand.Intn(def.ScreenWidth), -30)
				return e
			})
			b.Entities().Add(enemies1)
		case 1200:
			// Transition to the third phase after 20 seconds
			w.phase = 2
			for _, e := range b.Entities().Get(def.EntityTypeEnvironment) {
				if s, ok := e.(*environment.Spawner); ok {
					s.MarkAsRemovable()
				}
			}
			w.resetTicks()
		}
	case 2:
		switch w.ticksInPhase {
		case 0:
			// Switch from drifters to weavers
			enemies2 := environment.NewSpawner(w.densityEarly, func() def.Entity {
				e, _ := enemy.NewWeaver(rand.Intn(def.ScreenWidth), -30)
				return e
			})
			b.Entities().Add(enemies2)
		case 600:
			// Transition to the mine field
			w.phase = 3
			for _, e := range b.Entities().Get(def.EntityTypeEnvironment) {
				if s, ok := e.(*environment.Spawner); ok {
					s.MarkAsRemovable()
				} else if af, ok := e.(*environment.AsteroidField); ok {
					af.MarkAsRemovable()
				}
			}
			w.resetTicks()
		}
	case 3:
		if w.ticksInPhase == 0 {
			for i := range w.minefieldWidthEarly {
				for j := range w.minefieldHeightEarly {
					x := i * w.mineSpacingEarly
					y := j * w.mineSpacingEarly
					mine, _ := enemy.NewPathMine(x+10, y-b.Height(), w.minePathEarly)
					if w.minefieldDrift == 0 {
						w.minefieldDrift = mine.GetDrift()
					}
					b.Entities().Add(mine)
				}
			}
		}
		if w.minefieldDrift != 0 && w.ticksInPhase >= int(1.25*float64(b.Height())/w.minefieldDrift) {
			w.minefieldDrift = 0
			for _, e := range b.Entities().Get(def.EntityTypeEnemy) {
				if m, ok := e.(*enemy.PathMine); ok {
					m.SetDrift(w.minefieldDrift)
				}
			}
		}
		// Transition to next phase when the minefield has been cleared
		if len(b.Entities().Get(def.EntityTypeEnemy)) == 0 {
			w.phase = 4
			w.resetTicks()
		}
	}

	w.ticksInPhase++
}

func (w *Wave1) resetTicks() {
	w.ticksInPhase = -1
}

func (w *Wave1) Draw(img *ebit.Image) {}

func (w *Wave1) CanBeRemoved() bool {
	return false
}
