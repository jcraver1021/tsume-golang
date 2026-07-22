package def

// PreciseCollider is an optional interface for entities that need
// pixel-perfect or shape-based collision detection (narrow phase)
type PreciseCollider interface {
	Entity
	// CollidesWith performs precise collision detection
	// Only called after BoundingBoxOverlaps returns true
	CollidesWith(other Entity) bool
}

// Collides performs two-phase collision detection between entities
// Phase 1: Fast bounding box check via BoundingBoxOverlaps()
// Phase 2: Precise check via CollidesWith() if at least one implements PreciseCollider
// Returns true only if entities are actually colliding
func Collides(a, b Entity) bool {
	if !a.BoundingBoxOverlaps(b) {
		return false
	}

	preciseA, aHasPrecise := a.(PreciseCollider)
	preciseB, bHasPrecise := b.(PreciseCollider)

	if aHasPrecise {
		return preciseA.CollidesWith(b)
	}

	if bHasPrecise {
		return preciseB.CollidesWith(a)
	}

	return true
}
