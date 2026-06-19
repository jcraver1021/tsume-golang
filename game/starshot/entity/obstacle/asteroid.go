package obstacle

import (
	"image/color"
	"math/rand"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
)

// AsteroidSize defines the size category of an asteroid
type AsteroidSize int

const (
	AsteroidSmall AsteroidSize = iota
	AsteroidMedium
	AsteroidLarge
)

// Asteroid represents a space rock obstacle
type Asteroid struct {
	x, y          int
	width, height int
	speed         int
	size          AsteroidSize
	shape         [][]bool // Irregular shape for pixel rendering
	baseColor     color.RGBA
}

// NewAsteroid creates a new asteroid with the specified size at the given position
func NewAsteroid(x, y int, size AsteroidSize) *Asteroid {
	a := &Asteroid{
		x:    x,
		y:    y,
		size: size,
	}

	// Configure based on size
	switch size {
	case AsteroidSmall:
		a.width = 8
		a.height = 8
		a.speed = 3
		a.shape = generateSmallAsteroidShape()
	case AsteroidMedium:
		a.width = 16
		a.height = 16
		a.speed = 2
		a.shape = generateMediumAsteroidShape()
	case AsteroidLarge:
		a.width = 24
		a.height = 24
		a.speed = 1
		a.shape = generateLargeAsteroidShape()
	}

	// Randomize color slightly (brownish-gray)
	grayValue := uint8(100 + rand.Intn(50))
	a.baseColor = color.RGBA{
		R: grayValue + 20,
		G: grayValue,
		B: grayValue - 10,
		A: 255,
	}

	return a
}

// NewRandomAsteroid creates an asteroid with random size at the given position
func NewRandomAsteroid(x, y int) *Asteroid {
	// Weight toward smaller asteroids (60% small, 30% medium, 10% large)
	roll := rand.Float64()
	var size AsteroidSize
	if roll < 0.6 {
		size = AsteroidSmall
	} else if roll < 0.9 {
		size = AsteroidMedium
	} else {
		size = AsteroidLarge
	}
	return NewAsteroid(x, y, size)
}

func (a *Asteroid) Type() def.EntityType {
	return def.EntityTypeObstacle
}

func (a *Asteroid) Location() (x, y int) {
	return a.x, a.y
}

func (a *Asteroid) Dimensions() (width, height int) {
	return a.width, a.height
}

func (a *Asteroid) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(a.x+a.width < ox || a.x > ox+ow || a.y+a.height < oy || a.y > oy+oh)
}

// CollidesWith implements precise collision detection for irregular asteroid shape
func (a *Asteroid) CollidesWith(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()

	// Check each solid pixel in the asteroid shape
	for row := range a.shape {
		for col := range a.shape[row] {
			if a.shape[row][col] {
				px := a.x + col
				py := a.y + row

				// Check if this pixel overlaps with other entity's bounding box
				if px >= ox && px < ox+ow && py >= oy && py < oy+oh {
					return true
				}
			}
		}
	}
	return false
}

func (a *Asteroid) Act(b def.Scene) {
	// Move down the screen
	a.y += a.speed
}

func (a *Asteroid) Draw(img *ebit.Image) {
	for row := range a.shape {
		for col := range a.shape[row] {
			if a.shape[row][col] {
				img.Set(a.x+col, a.y+row, a.baseColor)
			}
		}
	}
}

func (a *Asteroid) CanBeRemoved() bool {
	// Remove when off bottom of screen
	return a.y > def.ScreenHeight
}

// generateSmallAsteroidShape creates an 8x8 irregular rock shape
func generateSmallAsteroidShape() [][]bool {
	return [][]bool{
		{false, false, true, true, true, false, false, false},
		{false, true, true, true, true, true, false, false},
		{true, true, true, true, true, true, true, false},
		{true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, false},
		{false, true, true, true, true, true, true, false},
		{false, false, true, true, true, true, false, false},
		{false, false, false, true, true, false, false, false},
	}
}

// generateMediumAsteroidShape creates a 16x16 irregular rock shape
func generateMediumAsteroidShape() [][]bool {
	return [][]bool{
		{false, false, false, true, true, true, true, true, true, false, false, false, false, false, false, false},
		{false, false, true, true, true, true, true, true, true, true, true, false, false, false, false, false},
		{false, true, true, true, true, true, true, true, true, true, true, true, false, false, false, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false},
		{false, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false},
		{false, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false},
		{false, false, true, true, true, true, true, true, true, true, true, true, false, false, false, false},
		{false, false, false, true, true, true, true, true, true, true, true, false, false, false, false, false},
		{false, false, false, false, true, true, true, true, true, true, false, false, false, false, false, false},
		{false, false, false, false, false, true, true, true, true, false, false, false, false, false, false, false},
	}
}

// generateLargeAsteroidShape creates a 24x24 irregular rock shape
func generateLargeAsteroidShape() [][]bool {
	return [][]bool{
		{false, false, false, false, false, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, true, true, true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false},
		{false, false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false, false, false, false, false},
		{false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false, false, false, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false, false, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false},
		{true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false},
		{false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false},
		{false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false, false},
		{false, false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false, false, false},
		{false, false, false, true, true, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false, false, false, false},
		{false, false, false, false, true, true, true, true, true, true, true, true, true, true, true, true, true, false, false, false, false, false, false, false},
		{false, false, false, false, false, true, true, true, true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, true, true, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, true, true, true, true, true, true, true, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, true, true, true, true, true, false, false, false, false, false, false, false, false, false, false, false},
		{false, false, false, false, false, false, false, false, false, true, true, true, false, false, false, false, false, false, false, false, false, false, false, false},
	}
}
