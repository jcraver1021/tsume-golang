package def

// Weapon is implemented by anything that can be equipped and fired by the player.
type Weapon interface {
	Fire(originX, originY int, scene Scene) // called to fire the weapon and spawn entities from the specified origin within the current scene
	TickCooldown()                          // called each frame to update the weapon's internal cooldown timer
	Ready() bool                            // returns true if the weapon is ready to fire (cooldown complete)
}

// AmmoBased is implemented by weapons that consume finite ammo rather than
// firing unlimitedly. Used by the HUD to display remaining ammo.
type AmmoBased interface {
	Ammo() int
	MaxAmmo() int
}

// Explosive is implemented by projectiles that detonate with area damage.
type Explosive interface {
	Entity
	BlastRadius() float64 // damage falloff distance in pixels
	BlastDamage() int     // flat HP removed from every Damageable entity whose center is within the blast radius
}

// SelfDetonating is an optional interface for entities that can trigger their own detonation based on internal state.
type SelfDetonating interface {
	ReadyToDetonate() bool // returns true if the entity should trigger its self-detonation this frame
}
