package projectile

import (
	"image/color"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
)

const (
	bulletWidth  = 3
	bulletHeight = 8
	bulletSpeed  = 10
)

type Bullet struct {
	x, y   int
	sprite *draw.ColorMatrix
	dead   bool
}

func NewBullet(x, y int) *Bullet {
	return &Bullet{
		x:      x,
		y:      y,
		sprite: generateBulletSprite(),
	}
}

func generateBulletSprite() *draw.ColorMatrix {
	// 3×8 vertical bolt: bright white core with cyan glow trail
	colors := draw.ColorMap{
		"0": {0, 0, 0, 0},         // transparent
		"1": {255, 255, 255, 255}, // white core
		"2": {80, 220, 255, 255},  // cyan mid
		"3": {40, 120, 200, 120},  // blue dim trail
	}

	// Row layout: top = bright, bottom = dim trail
	rows := []string{
		"010",
		"111",
		"111",
		"111",
		"222",
		"222",
		"233",
		"030",
	}

	matrix := make([][]draw.ColorKey, len(rows))
	for r, row := range rows {
		matrix[r] = make([]draw.ColorKey, len(row))
		for c, ch := range row {
			matrix[r][c] = draw.ColorKey(string(ch))
		}
	}

	cm, err := draw.NewColorMatrix(matrix, &colors, nil)
	if err != nil {
		// Fallback: single white pixel column
		fb := make([][]draw.ColorKey, bulletHeight)
		fbc := draw.ColorMap{"1": {255, 255, 255, 255}, "0": {0, 0, 0, 0}}
		for r := range fb {
			fb[r] = []draw.ColorKey{"0", "1", "0"}
		}
		cm, _ = draw.NewColorMatrix(fb, &fbc, nil)
	}
	return cm
}

func (b *Bullet) Type() def.EntityType {
	return def.EntityTypeTeam
}

func (b *Bullet) Location() (int, int) {
	return b.x, b.y
}

func (b *Bullet) Dimensions() (int, int) {
	return bulletWidth, bulletHeight
}

func (b *Bullet) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(b.x+bulletWidth < ox || b.x > ox+ow || b.y+bulletHeight < oy || b.y > oy+oh)
}

func (b *Bullet) Act(scene def.Scene) {
	b.y -= bulletSpeed
}

func (b *Bullet) Draw(img *ebit.Image) {
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

func (b *Bullet) CanBeRemoved() bool {
	return b.dead || b.y+bulletHeight < 0
}

// MarkDestroyed removes the bullet (called on collision).
func (b *Bullet) MarkDestroyed() {
	b.dead = true
}

// ImpactColor returns a color for a small flash effect (unused for now).
func (b *Bullet) ImpactColor() color.RGBA {
	return color.RGBA{80, 220, 255, 255}
}
