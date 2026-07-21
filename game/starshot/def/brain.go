package def

// SignalKind identifies what an entity's sensor detected.
type SignalKind int

const (
	SignalSelf     SignalKind = iota // the perceiving entity itself
	SignalPlayer                     // the player ship
	SignalAlly                       // a friendly entity
	SignalObstacle                   // a physical hazard in the flight path
	SignalDanger                     // area-of-effect threat (blast radius, etc.)
)

// Condition is the perceived health state of an entity.
// Thresholds are set by the sensor (Perceive), not exposed as raw HP.
type Condition int

const (
	ConditionHealthy  Condition = iota // > 50% HP
	ConditionDamaged                   // ≤ 50% HP
	ConditionCritical                  // ≤ 25% HP
)

// ConditionFor maps current/max HP to a Condition.
// Centralizing the thresholds keeps sensors consistent across entity types.
func ConditionFor(current, max int) Condition {
	if max == 0 {
		return ConditionCritical
	}
	switch pct := float64(current) / float64(max); {
	case pct > 0.5:
		return ConditionHealthy
	case pct > 0.25:
		return ConditionDamaged
	default:
		return ConditionCritical
	}
}

// Signal is a single thing an entity's sensor detected.
// Direction is a normalized vector from the perceiver toward the source;
// zero for SignalSelf (you are not in a direction relative to yourself).
// Distance is pixels to the source center; zero for SignalSelf.
// Condition is meaningful for living sources (Self, Player, Ally);
// zero for non-living signals (Obstacle, Danger).
type Signal struct {
	Kind      SignalKind
	Direction [2]float64
	Distance  float64
	Condition Condition
}

// Perception is the complete set of signals an entity's sensor produced this frame.
// The brain receives this and decides what to do — it never sees the raw scene.
type Perception []Signal

// Intent is what an entity wants to do this frame.
// Direction is a normalized heading vector; Speed is the magnitude in pixels/frame.
// A zero Direction means no movement.
// Fire signals that the entity wants to shoot; FireAim is the normalized direction
// to fire in. The entity applies its own rate-limiting before spawning a projectile.
type Intent struct {
	Direction [2]float64
	Speed     float64
	Fire      bool
	FireAim   [2]float64
}

// Brain decides what an entity should do given its current Perception.
// Implementations range from hardcoded state machines to external ML models.
// The entity calls Decide each frame after building its Perception from the scene.
type Brain interface {
	Decide(Perception) Intent
}
