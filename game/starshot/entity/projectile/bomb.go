package projectile

import (
	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
	"tsumegolang/game/starshot/entity/effects"
)

const (
	bombWidth       = 12
	bombHeight      = 13
	bombSpeed       = 3
	BombBlastRadius = 64.0
	BombBlastDamage = 999 // one-shots any asteroid on a direct hit
)

// Bomb is a slow heavy projectile that detonates in an area on contact.
type Bomb struct {
	x, y   int
	sprite *draw.ColorMatrix
	dead   bool
}

func NewBomb(x, y int) *Bomb {
	return &Bomb{
		x:      x - bombWidth/2,
		y:      y,
		sprite: generateBombSprite(),
	}
}

func generateBombSprite() *draw.ColorMatrix {
	colors := draw.ColorMap{
		"0": {0, 0, 0, 0},
		"1": {37, 37, 37, 255},    // dark outer iron
		"2": {66, 66, 66, 255},    // iron body
		"3": {106, 106, 106, 255}, // iron surface
		"4": {136, 51, 34, 255},   // warm inner ring
		"5": {221, 85, 0, 255},    // orange inner glow
		"6": {255, 153, 0, 255},   // hot core
		"7": {255, 238, 0, 255},   // fuse spark
	}

	rows := []string{
		"001111111100",
		"011222222110",
		"122333333221",
		"123444444321",
		"124455554421",
		"124556655421",
		"124455554421",
		"123444444321",
		"122333333221",
		"011222222110",
		"001111111100",
		"000033000000",
		"000037300000",
	}

	matrix := make([][]draw.ColorKey, len(rows))
	for r, row := range rows {
		matrix[r] = make([]draw.ColorKey, len(row))
		for c, ch := range row {
			matrix[r][c] = draw.ColorKey(string(ch))
		}
	}

	cm, _ := draw.NewColorMatrix(matrix, &colors, nil)
	return cm
}

func (b *Bomb) Type() def.EntityType {
	return def.EntityTypeTeam
}

func (b *Bomb) Location() (int, int) {
	return b.x, b.y
}

func (b *Bomb) Dimensions() (int, int) {
	return bombWidth, bombHeight
}

func (b *Bomb) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(b.x+bombWidth < ox || b.x > ox+ow || b.y+bombHeight < oy || b.y > oy+oh)
}

func (b *Bomb) Act(_ def.Scene) {
	b.y -= bombSpeed
}

func (b *Bomb) Draw(img *ebit.Image) {
	if b.dead {
		return
	}
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

func (b *Bomb) CanBeRemoved() bool {
	return b.dead || b.y+bombHeight < 0
}

func (b *Bomb) MarkDestroyed() {
	b.dead = true
}

// Mortal — the bomb gets a visual explosion at detonation point.

func (b *Bomb) GetDeathEffect() def.DeathEffect {
	return def.DeathEffect{
		SpawnVisualEffect: func(cx, cy int, scene def.Scene) {
			if exp, err := effects.NewExplosion(cx, cy, effects.ExplosionLarge); err == nil {
				scene.Entities().Add(exp)
			}
		},
	}
}

func (b *Bomb) MarkAsDead(_ def.Scene) {
	b.dead = true
}

func (b *Bomb) IsDead() bool {
	return b.dead
}

// Explosive

func (b *Bomb) BlastRadius() float64 {
	return BombBlastRadius
}

func (b *Bomb) BlastDamage() int {
	return BombBlastDamage
}
