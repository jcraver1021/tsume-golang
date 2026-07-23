package player

import (
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
	"tsumegolang/game/starshot/entity/projectile"
)

const basicGunCooldownFrames = 12 // ~5 shots/second at 60 FPS

// Gun is the common struct for all forward-firing player weapons.
// Different gun variants are created by different constructors; firing
// behavior varies via the fireFunc field so new types need only a new
// constructor and a sprite — no new struct required.
type Gun struct {
	sprite         *draw.ColorMatrix
	cooldown       int
	cooldownFrames int
	mountY         int
	fireFunc       func(x, y int, scene def.Scene)
}

// NewBasicGun returns a rapid-fire centered bullet gun.
func NewBasicGun() (*Gun, error) {
	data, err := spriteFiles.ReadFile("sprites/gun_basic.yaml")
	if err != nil {
		return nil, err
	}

	sprite, err := draw.ColorMatrixFromBytes(data)
	if err != nil {
		return nil, err
	}

	return &Gun{
		sprite:         sprite,
		cooldownFrames: basicGunCooldownFrames,
		mountY:         0,
		fireFunc: func(x, y int, scene def.Scene) {
			scene.Entities().Add(projectile.NewBullet(x, y))
		},
	}, nil
}

// Add more gun types here as needed

func (g *Gun) TickCooldown() {
	if g.cooldown > 0 {
		g.cooldown--
	}
}

func (g *Gun) Ready() bool {
	return g.cooldown == 0
}

func (g *Gun) Fire(originX, originY int, scene def.Scene) {
	g.fireFunc(originX, originY, scene)
	g.cooldown = g.cooldownFrames
}

func (g *Gun) Sprite() *draw.ColorMatrix {
	return g.sprite
}

func (g *Gun) MountOffsetX(hullWidth int) int {
	return (hullWidth - g.sprite.Width()) / 2
}

func (g *Gun) MountOffsetY() int {
	return g.mountY
}
