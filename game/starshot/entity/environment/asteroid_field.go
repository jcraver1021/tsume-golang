package environment

import (
	"math/rand"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/obstacle"
)

type AsteroidField struct {
	asteroidDensity float64
	sizeFn          func() obstacle.AsteroidSize
	canBeRemoved    bool
}

func NewAsteroidField(density float64, sizeFn func() obstacle.AsteroidSize) *AsteroidField {
	return &AsteroidField{
		asteroidDensity: density,
		sizeFn:          sizeFn,
		canBeRemoved:    false,
	}
}

func (af *AsteroidField) Type() def.EntityType {
	return def.EntityTypeEnvironment
}

func (af *AsteroidField) Onscreen() def.OnScreen {
	return def.OffScreen
}

func (af *AsteroidField) Location() (x, y int) {
	return 0, 0
}

func (af *AsteroidField) Dimensions() (width, height int) {
	return def.ScreenWidth, def.ScreenHeight
}

func (af *AsteroidField) BoundingBoxOverlaps(other def.Entity) bool {
	return false
}

func (af *AsteroidField) maybeAddAsteroid(b def.Scene) *obstacle.Asteroid {
	if rand.Float64() < af.asteroidDensity {
		size := af.sizeFn()
		x := rand.Intn(b.Width())
		_, h := size.Dimensions()
		y := -rand.Intn(h) // start above the screen
		return obstacle.NewAsteroid(x, y, size)
	}

	return nil
}

func (af *AsteroidField) Act(b def.Scene) {
	if asteroid := af.maybeAddAsteroid(b); asteroid != nil {
		b.Entities().Add(asteroid)
	}
}

func (af *AsteroidField) Draw(img *ebit.Image) {}

func (af *AsteroidField) CanBeRemoved() bool {
	return af.canBeRemoved
}

func (af *AsteroidField) MarkAsRemovable() {
	af.canBeRemoved = true
}
