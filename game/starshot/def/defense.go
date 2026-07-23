package def

// Mortal is an optional interface for entities that can die and spawn effects
type Mortal interface {
	Entity
	GetDeathEffect() DeathEffect
	MarkAsDead(scene Scene)
	IsDead() bool
}

// DeathEffect specifies what happens when an entity dies.
// The game calls SpawnVisualEffect when the entity dies, if it is non-nil.
// It should handle any visual effects associated with the entity's death.
// Further game effects (e.g. explosion damage) are handled elsewhere
// Slowdown is mainly for player death, but can also be used for particularly impactful entity deaths.
type DeathEffect struct {
	SpawnVisualEffect  func(cx, cy int, scene Scene) // nil = no visual effect
	SlowdownMultiplier float64                       // 0.0 = no slowdown, 0.3 = 30% speed, 1.0 = normal
	SlowdownDuration   int                           // Frames to maintain slowdown (0 = no slowdown)
}

// Damageable is an optional interface for entities with hit points.
// TakeDamage reduces HP; callers should then check IsDead() via Mortal.
type Damageable interface {
	Entity
	TakeDamage(amount int)
	CurrentHP() int
	MaxHP() int
}

// Impulsable is an optional interface for entities that can receive
// a velocity impulse from projectile hits.
type Impulsable interface {
	Entity
	ApplyImpulse(dvx, dvy float64)
}
