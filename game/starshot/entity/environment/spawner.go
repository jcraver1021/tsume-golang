package environment

import (
	"math/rand"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
)

type Spawner struct {
	enemyDensity float64
	enemyFn      func() def.Entity
	canBeRemoved bool
}

func NewSpawner(density float64, enemyFn func() def.Entity) *Spawner {
	return &Spawner{
		enemyDensity: density,
		enemyFn:      enemyFn,
		canBeRemoved: false,
	}
}

func (s *Spawner) Type() def.EntityType {
	return def.EntityTypeEnvironment
}

func (s *Spawner) Onscreen() def.OnScreen {
	return def.OffScreen
}

func (s *Spawner) Location() (x, y int) {
	return 0, 0
}

func (s *Spawner) Dimensions() (width, height int) {
	return def.ScreenWidth, def.ScreenHeight
}

func (s *Spawner) BoundingBoxOverlaps(other def.Entity) bool {
	return false
}

func (s *Spawner) Act(b def.Scene) {
	if rand.Float64() < s.enemyDensity {
		enemy := s.enemyFn() // enemy function should take care of placement
		if enemy != nil {
			b.Entities().Add(enemy)
		}
	}
}

func (s *Spawner) Draw(img *ebit.Image) {}

func (s *Spawner) CanBeRemoved() bool {
	return s.canBeRemoved
}

func (s *Spawner) MarkAsRemovable() {
	s.canBeRemoved = true
}
