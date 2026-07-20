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
	// Broad phase: cheap bounding box check
	if !a.BoundingBoxOverlaps(b) {
		return false
	}

	// Narrow phase: if either entity has precise collision, use it
	preciseA, aHasPrecise := a.(PreciseCollider)
	preciseB, bHasPrecise := b.(PreciseCollider)

	if aHasPrecise {
		// A has precise collision - use it
		return preciseA.CollidesWith(b)
	}

	if bHasPrecise {
		// B has precise collision - use it
		return preciseB.CollidesWith(a)
	}

	// Neither has precise collision - bounding box overlap is sufficient
	return true
}
