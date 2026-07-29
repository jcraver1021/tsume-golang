package projectile

import (
	"image/color"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/pkg/tool/sprite"
)

const (
	bulletWidth  = 3
	bulletHeight = 8
	bulletSpeed  = 10
)

type Bullet struct {
	x, y   int
	sprite *sprite.Sprite
	dead   bool
}

func NewBullet(x, y int) *Bullet {
	return &Bullet{
		x:      x,
		y:      y,
		sprite: generateBulletSprite(),
	}
}

func generateBulletSprite() *sprite.Sprite {
	// 3×8 vertical bolt: bright white core with cyan glow trail
	palette := sprite.NewPalette()
	key0, _ := palette.Add(color.RGBA{0, 0, 0, 0})         // transparent
	key1, _ := palette.Add(color.RGBA{255, 255, 255, 255}) // white core
	key2, _ := palette.Add(color.RGBA{80, 220, 255, 255})  // cyan mid
	key3, _ := palette.Add(color.RGBA{40, 120, 200, 120})  // blue dim trail

	keyMap := map[rune]sprite.ColorKey{
		'0': key0,
		'1': key1,
		'2': key2,
		'3': key3,
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

	matrix := make([][]sprite.ColorKey, len(rows))
	for r, row := range rows {
		matrix[r] = make([]sprite.ColorKey, len(row))
		for c, ch := range row {
			matrix[r][c] = keyMap[ch]
		}
	}

	s, err := sprite.NewSprite(matrix, palette, map[sprite.ColorKey]*sprite.AnimationSequence{})
	if err != nil {
		// Fallback: single white pixel column
		fbPalette := sprite.NewPalette()
		fbKey1, _ := fbPalette.Add(color.RGBA{255, 255, 255, 255})
		fbKey0, _ := fbPalette.Add(color.RGBA{0, 0, 0, 0})
		fb := make([][]sprite.ColorKey, bulletHeight)
		for r := range fb {
			fb[r] = []sprite.ColorKey{fbKey0, fbKey1, fbKey0}
		}
		s, _ = sprite.NewSprite(fb, fbPalette, map[sprite.ColorKey]*sprite.AnimationSequence{})
	}
	return s
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
