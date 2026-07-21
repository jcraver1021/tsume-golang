package def

// Weapon is implemented by anything that can be equipped and fired by the player.
// Fire is called with the spawn origin and the current scene; the weapon is
// responsible for creating and adding its projectiles.
// TickCooldown and Ready decouple rate-of-fire from the player's own update loop.
type Weapon interface {
	Fire(originX, originY int, scene Scene)
	TickCooldown()
	Ready() bool
}

// AmmoBased is implemented by weapons that consume finite ammo rather than
// firing unlimitedly. Used by the HUD to display remaining ammo.
type AmmoBased interface {
	Ammo() int
	MaxAmmo() int
}

// Explosive is implemented by projectiles that detonate with area damage.
// BlastRadius is the damage falloff distance in pixels; BlastDamage is the
// flat HP removed from every Damageable entity whose center is within that radius.
type Explosive interface {
	Entity
	BlastRadius() float64
	BlastDamage() int
}

// SelfDetonating is implemented by entities that trigger their own detonation
// based on internal state (e.g. a proximity timer). game.go checks ReadyToDetonate
// each frame and calls handleDeath when it returns true.
type SelfDetonating interface {
	ReadyToDetonate() bool
}
