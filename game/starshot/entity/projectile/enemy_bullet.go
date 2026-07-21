package projectile

import (
	"math"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
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
	sprite *draw.ColorMatrix
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

func generatePhotonTorpedoSprite() *draw.ColorMatrix {
	// Color codes for static zones and animation frame colors.
	// 'p' and 'q' are the core animation frame colors, kept out of animSeqs.
	colorCodes := draw.ColorMap{
		"0": {0, 0, 0, 0},         // transparent
		"o": {160, 10, 0, 180},    // outer halo: deep red, semi-transparent
		"m": {220, 50, 0, 230},    // mid ring: orange-red
		"i": {255, 140, 0, 255},   // inner ring: bright orange
		"p": {255, 240, 80, 255},  // core frame A: hot yellow
		"q": {255, 255, 220, 255}, // core frame B: near-white
	}

	// Core pulses between hot yellow and near-white every 6 frames.
	coreAnim := draw.NewAnimationSequence(
		&colorCodes,
		[]draw.ColorKey{"p", "p", "q", "p"},
		6,
	)
	animSeqs := map[draw.ColorKey]*draw.AnimationSequence{"c": coreAnim}

	const center = float64(torpedoSize-1) / 2
	matrix := make([][]draw.ColorKey, torpedoSize)
	for r := range matrix {
		matrix[r] = make([]draw.ColorKey, torpedoSize)
		for c := range matrix[r] {
			dx := float64(c) - center
			dy := float64(r) - center
			dist := math.Sqrt(dx*dx + dy*dy)
			switch {
			case dist < 1.6:
				matrix[r][c] = "c" // animated core
			case dist < 2.8:
				matrix[r][c] = "i"
			case dist < 4.0:
				matrix[r][c] = "m"
			case dist < torpedoRadius:
				matrix[r][c] = "o"
			default:
				matrix[r][c] = "0"
			}
		}
	}

	cm, err := draw.NewColorMatrix(matrix, &colorCodes, animSeqs)
	if err != nil {
		// Fallback: single red pixel
		fb := [][]draw.ColorKey{{"r"}}
		fbc := draw.ColorMap{"r": {255, 50, 0, 255}}
		cm, _ = draw.NewColorMatrix(fb, &fbc, nil)
	}
	return cm
}

func (b *EnemyBullet) Type() def.EntityType { return def.EntityTypeEnemyTeam }

func (b *EnemyBullet) Location() (int, int) { return b.x, b.y }

func (b *EnemyBullet) Dimensions() (int, int) { return torpedoSize, torpedoSize }

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

func (b *EnemyBullet) MarkDestroyed() { b.dead = true }
