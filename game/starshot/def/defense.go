package def

// ExplosionSize specifies the visual scale of an explosion effect
type ExplosionSize int

const (
	ExplosionSmall ExplosionSize = iota
	ExplosionMedium
	ExplosionLarge
)

// Mortal is an optional interface for entities that can die and spawn effects
type Mortal interface {
	Entity
	GetDeathEffect() DeathEffect
	MarkAsDead(scene Scene)
	IsDead() bool
}

// DeathEffect specifies what happens when an entity dies
type DeathEffect struct {
	ExplosionSize      ExplosionSize
	SlowdownMultiplier float64 // 0.0 = no slowdown, 0.3 = 30% speed, 1.0 = normal
	SlowdownDuration   int     // Frames to maintain slowdown (0 = no slowdown)
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
