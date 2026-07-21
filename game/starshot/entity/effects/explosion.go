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

// ExplosionSize specifies the visual scale of an explosion effect.
type ExplosionSize int

const (
	ExplosionSmall  ExplosionSize = iota
	ExplosionMedium ExplosionSize = iota
	ExplosionLarge  ExplosionSize = iota
)

// LoadExplosionSprite loads the sprite sheet for the given explosion size.
func LoadExplosionSprite(size ExplosionSize) (*draw.ColorMatrix, error) {
	var spriteFile string
	switch size {
	case ExplosionSmall:
		spriteFile = "sprites/explosion_small.yaml"
	case ExplosionMedium:
		spriteFile = "sprites/explosion_medium.yaml"
	case ExplosionLarge:
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
	cachedImg     *ebit.Image
	pixelBuf      []byte
	drawScale     float64
	frameCount    int
	maxFrames     int
}

// NewExplosion creates an explosion entity at the given center coordinates.
func NewExplosion(cx, cy int, size ExplosionSize) (*Explosion, error) {
	return NewExplosionScaled(cx, cy, size, 1.0)
}

// NewExplosionScaled creates an explosion drawn at scale times its natural sprite size.
// Use this to match the visual footprint to a blast radius: scale = blastDiameter / spriteWidth.
func NewExplosionScaled(cx, cy int, size ExplosionSize, scale float64) (*Explosion, error) {
	sprite, err := LoadExplosionSprite(size)
	if err != nil {
		return nil, err
	}

	var maxFrames int
	switch size {
	case ExplosionSmall:
		maxFrames = 40 // 4 frames × 10 ticks/frame
	case ExplosionMedium:
		maxFrames = 60 // 6 frames × 10 ticks/frame
	case ExplosionLarge:
		maxFrames = 96 // 8 frames × 12 ticks/frame
	default:
		return nil, fmt.Errorf("unknown explosion size: %d", size)
	}

	naturalW, naturalH := sprite.Dimensions()
	drawnW := int(float64(naturalW) * scale)
	drawnH := int(float64(naturalH) * scale)

	return &Explosion{
		x:         cx - drawnW/2,
		y:         cy - drawnH/2,
		width:     drawnW,
		height:    drawnH,
		sprite:    sprite,
		cachedImg: ebit.NewImage(naturalW, naturalH),
		pixelBuf:  make([]byte, naturalW*naturalH*4),
		drawScale: scale,
		maxFrames: maxFrames,
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
	draw.DrawScaled(img, e.cachedImg, e.pixelBuf, e.sprite, float64(e.x), float64(e.y), e.drawScale)
}

func (e *Explosion) CanBeRemoved() bool {
	return e.frameCount >= e.maxFrames
}
