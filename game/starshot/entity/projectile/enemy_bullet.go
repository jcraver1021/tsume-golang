package projectile

import (
	"image/color"
	"math"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/pkg/tool/sprite"
)

const (
	torpedoSize   = 11
	torpedoRadius = 5.0
	torpedoSpeed  = 4.0

	EnemyBulletDamage = 1
)

// EnemyBullet is a directional photon-torpedo projectile fired by enemy entities.
// It travels in an arbitrary direction set at spawn time and pulses with a
// red-yellow glow each frame via an animated core.
type EnemyBullet struct {
	fx, fy float64
	x, y   int
	vx, vy float64
	dead   bool
	sprite *sprite.Sprite
}

// NewEnemyBullet spawns a torpedo centered at (cx, cy) traveling in aim direction.
func NewEnemyBullet(cx, cy int, aim [2]float64) *EnemyBullet {
	half := torpedoSize / 2
	return &EnemyBullet{
		fx:     float64(cx - half),
		fy:     float64(cy - half),
		x:      cx - half,
		y:      cy - half,
		vx:     aim[0] * torpedoSpeed,
		vy:     aim[1] * torpedoSpeed,
		sprite: generatePhotonTorpedoSprite(),
	}
}

func generatePhotonTorpedoSprite() *sprite.Sprite {
	palette := sprite.NewPalette()
	transparent, _ := palette.Add(color.RGBA{0, 0, 0, 0})
	outer, _ := palette.Add(color.RGBA{160, 10, 0, 180})
	mid, _ := palette.Add(color.RGBA{220, 50, 0, 230})
	inner, _ := palette.Add(color.RGBA{255, 140, 0, 255})
	frameA, _ := palette.Add(color.RGBA{255, 240, 80, 255})
	frameB, _ := palette.Add(color.RGBA{255, 255, 220, 255})
	coreKey, _ := palette.Reserve("c") // reserve a key for the animation slot

	seq, err := sprite.NewAnimationSequence(palette, []sprite.ColorKey{frameA, frameA, frameB, frameA}, 6)
	if err != nil {
		// This should not fail with valid inputs; fall back to static sprite
		return generatePhotonTorpedoFallback()
	}
	animSeqs := map[sprite.ColorKey]*sprite.AnimationSequence{coreKey: seq}

	const center = float64(torpedoSize-1) / 2
	matrix := make([][]sprite.ColorKey, torpedoSize)
	for r := range matrix {
		matrix[r] = make([]sprite.ColorKey, torpedoSize)
		for c := range matrix[r] {
			dx := float64(c) - center
			dy := float64(r) - center
			dist := math.Sqrt(dx*dx + dy*dy)
			switch {
			case dist < 1.6:
				matrix[r][c] = coreKey // animated core
			case dist < 2.8:
				matrix[r][c] = inner
			case dist < 4.0:
				matrix[r][c] = mid
			case dist < torpedoRadius:
				matrix[r][c] = outer
			default:
				matrix[r][c] = transparent
			}
		}
	}

	s, err := sprite.NewSprite(matrix, palette, animSeqs)
	if err != nil {
		return generatePhotonTorpedoFallback()
	}
	return s
}

func generatePhotonTorpedoFallback() *sprite.Sprite {
	fbPalette := sprite.NewPalette()
	fbKey, _ := fbPalette.Add(color.RGBA{255, 50, 0, 255})
	fb := [][]sprite.ColorKey{{fbKey}}
	s, _ := sprite.NewSprite(fb, fbPalette, map[sprite.ColorKey]*sprite.AnimationSequence{})
	return s
}

func (b *EnemyBullet) Type() def.EntityType {
	return def.EntityTypeEnemyTeam
}

func (b *EnemyBullet) Location() (int, int) {
	return b.x, b.y
}

func (b *EnemyBullet) Dimensions() (int, int) {
	return torpedoSize, torpedoSize
}

func (b *EnemyBullet) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(b.x+torpedoSize < ox || b.x > ox+ow || b.y+torpedoSize < oy || b.y > oy+oh)
}

func (b *EnemyBullet) Act(_ def.Scene) {
	b.fx += b.vx
	b.fy += b.vy
	b.x = int(b.fx)
	b.y = int(b.fy)
	b.sprite.Advance()
}

func (b *EnemyBullet) Draw(img *ebit.Image) {
	pixels := b.sprite.Render()
	for row := range pixels {
		for col := range pixels[row] {
			c := pixels[row][col]
			if c.A > 0 {
				img.Set(b.x+col, b.y+row, c)
			}
		}
	}
}

func (b *EnemyBullet) CanBeRemoved() bool {
	if b.dead {
		return true
	}
	return b.x+torpedoSize < 0 ||
		b.x > def.ScreenWidth ||
		b.y+torpedoSize < 0 ||
		b.y > def.ScreenHeight
}

func (b *EnemyBullet) MarkDestroyed() {
	b.dead = true
}
