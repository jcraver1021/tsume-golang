package projectile

import "tsumegolang/game/starshot/def"

const basicGunCooldownFrames = 12 // ~5 shots/second at 60 FPS

// BasicGun fires a single centered bullet upward.
type BasicGun struct {
	cooldown int
}

func NewBasicGun() *BasicGun {
	return &BasicGun{}
}

func (g *BasicGun) TickCooldown() {
	if g.cooldown > 0 {
		g.cooldown--
	}
}

func (g *BasicGun) Ready() bool {
	return g.cooldown == 0
}

func (g *BasicGun) Fire(originX, originY int, scene def.Scene) {
	scene.Entities().Add(NewBullet(originX, originY))
	g.cooldown = basicGunCooldownFrames
}
