package player

import (
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/projectile"
	"tsumegolang/pkg/tool/sprite"
)

const basicGunCooldownFrames = 12 // ~5 shots/second at 60 FPS

// Gun is the common struct for all forward-firing player weapons.
// Different gun variants are created by different constructors; firing
// behavior varies via the fireFunc field so new types need only a new
// constructor and a sprite — no new struct required.
type Gun struct {
	sprite         *sprite.Sprite
	cooldown       int
	cooldownFrames int
	mountY         int
	fireFunc       func(x, y int, scene def.Scene)
}

func NewBasicGun() (*Gun, error) {
	data, err := spriteFiles.ReadFile("sprites/gun_basic.yaml")
	if err != nil {
		return nil, err
	}

	s, err := sprite.SpriteFromBytes(data)
	if err != nil {
		return nil, err
	}

	return &Gun{
		sprite:         s,
		cooldownFrames: basicGunCooldownFrames,
		mountY:         0,
		fireFunc: func(x, y int, scene def.Scene) {
			scene.Entities().Add(projectile.NewBullet(x, y))
		},
	}, nil
}

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

func (g *Gun) Sprite() *sprite.Sprite {
	return g.sprite
}

func (g *Gun) MountOffsetX(hullWidth int) int {
	return (hullWidth - g.sprite.Width()) / 2
}

func (g *Gun) MountOffsetY() int {
	return g.mountY
}
