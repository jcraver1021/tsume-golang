package def

// SignalCategory identifies the categorization of a detected signal.
type SignalCategory int

const (
	SignalSelf     SignalCategory = iota // the perceiving entity itself
	SignalPlayer                         // the player ship
	SignalAlly                           // a friendly entity
	SignalObstacle                       // a physical hazard in the flight path
	SignalDanger                         // area-of-effect threat (blast radius, etc.)
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

// Signal represents a single detection.
type Signal struct {
	Kind      SignalCategory
	Direction [2]float64 // normalized vector from the perceiver toward the source (zero for self)
	Distance  float64    // distance to the source center (zero for self)
	Condition Condition  // meaningful for living sources (Self, Player, Ally); zero for non-living signals (Obstacle, Danger)
}

// Perception represents the complete set of signals detected by an entity in a signal frame.
// The "brain" does not see the raw state, only what their perception provides.
type Perception []Signal

// Intent represents the desired actions of an entity for the current frame.
type Intent struct {
	Direction [2]float64 // normalized heading vector; zero means no movement
	Speed     float64    // magnitude in pixels/frame
	Fire      bool       // signals that the entity wants to shoot
	FireAim   [2]float64 // normalized direction to fire in
}

// Brain decides the actions of an entity based on its Perception (called once per frame).
type Brain interface {
	Decide(Perception) Intent
}
