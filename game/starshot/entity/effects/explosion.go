package effects

import (
	"embed"
	"fmt"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
)

//go:embed sprites/*.yaml
var spriteFiles embed.FS

// LoadExplosionSprite loads an explosion sprite of the given size
// This is exported so other packages (like player) can load explosion sprites
// without duplicating the sprite files
func LoadExplosionSprite(size def.ExplosionSize) (*draw.ColorMatrix, error) {
	var spriteFile string
	switch size {
	case def.ExplosionSmall:
		spriteFile = "sprites/explosion_small.yaml"
	case def.ExplosionMedium:
		spriteFile = "sprites/explosion_medium.yaml"
	case def.ExplosionLarge:
		spriteFile = "sprites/explosion_large.yaml"
	default:
		return nil, fmt.Errorf("unknown explosion size: %d", size)
	}

	spriteData, err := spriteFiles.ReadFile(spriteFile)
	if err != nil {
		return nil, err
	}

	return draw.ColorMatrixFromBytes(spriteData)
}

type Explosion struct {
	x, y          int
	width, height int
	sprite        *draw.ColorMatrix
	frameCount    int
	maxFrames     int
}

// NewExplosion creates an explosion effect at the given location
func NewExplosion(x, y int, scene def.Scene, size def.ExplosionSize) (*Explosion, error) {
	// Load sprite using shared loader
	sprite, err := LoadExplosionSprite(size)
	if err != nil {
		return nil, err
	}

	// Determine duration based on size
	var maxFrames int
	switch size {
	case def.ExplosionSmall:
		maxFrames = 40 // 4 frames × 10 ticks/frame
	case def.ExplosionMedium:
		maxFrames = 60 // 6 frames × 10 ticks/frame
	case def.ExplosionLarge:
		maxFrames = 96 // 8 frames × 12 ticks/frame
	default:
		return nil, fmt.Errorf("unknown explosion size: %d", size)
	}

	width, height := sprite.Dimensions()

	return &Explosion{
		x:          x - width/2, // Center on spawn location
		y:          y - height/2,
		width:      width,
		height:     height,
		sprite:     sprite,
		frameCount: 0,
		maxFrames:  maxFrames,
	}, nil
}

func (e *Explosion) Type() def.EntityType {
	return def.EntityTypeEnvironment
}

func (e *Explosion) Location() (x, y int) {
	return e.x, e.y
}

func (e *Explosion) Dimensions() (width, height int) {
	return e.width, e.height
}

func (e *Explosion) BoundingBoxOverlaps(other def.Entity) bool {
	return false
}

func (e *Explosion) Act(scene def.Scene) {
	e.frameCount++
}

func (e *Explosion) Draw(img *ebit.Image) {
	pixels := e.sprite.Render()

	for row := range pixels {
		for col := range pixels[row] {
			color := pixels[row][col]
			if color.A > 0 {
				img.Set(e.x+col, e.y+row, color)
			}
		}
	}
}

func (e *Explosion) CanBeRemoved() bool {
	return e.frameCount >= e.maxFrames
}
