package def_test

import (
	"testing"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/testutil"
)

// overlapping returns two mock entities whose bounding boxes overlap.
func overlapping() (*testutil.MockEntity, *testutil.MockEntity) {
	a := testutil.NewMockEntity(def.EntityTypePlayer)
	a.X, a.Y, a.Width, a.Height = 0, 0, 20, 20

	b := testutil.NewMockEntity(def.EntityTypeEnemy)
	b.X, b.Y, b.Width, b.Height = 10, 10, 20, 20

	return a, b
}

// separated returns two mock entities whose bounding boxes do not overlap.
func separated() (*testutil.MockEntity, *testutil.MockEntity) {
	a := testutil.NewMockEntity(def.EntityTypePlayer)
	a.X, a.Y, a.Width, a.Height = 0, 0, 10, 10

	b := testutil.NewMockEntity(def.EntityTypeEnemy)
	b.X, b.Y, b.Width, b.Height = 100, 100, 10, 10

	return a, b
}

// --- Broad-phase only (no PreciseCollider) ---

func TestCollidesReturnsTrueWhenBoundingBoxesOverlap(t *testing.T) {
	a, b := overlapping()
	if !def.Collides(a, b) {
		t.Error("Collides = false, want true for overlapping bounding boxes")
	}
}

func TestCollidesReturnsFalseWhenBoundingBoxesSeparated(t *testing.T) {
	a, b := separated()
	if def.Collides(a, b) {
		t.Error("Collides = true, want false for non-overlapping bounding boxes")
	}
}

func TestCollidesIsSymmetric(t *testing.T) {
	a, b := overlapping()
	if def.Collides(a, b) != def.Collides(b, a) {
		t.Error("Collides(a,b) != Collides(b,a): collision must be symmetric")
	}
}

func TestCollidesTouchingEdgeIsCollision(t *testing.T) {
	// Right edge of a at x=10; left edge of b at x=10. Adjacent, not overlapping.
	a := testutil.NewMockEntity(def.EntityTypePlayer)
	a.X, a.Y, a.Width, a.Height = 0, 0, 10, 10

	b := testutil.NewMockEntity(def.EntityTypeEnemy)
	b.X, b.Y, b.Width, b.Height = 10, 0, 10, 10

	// BoundingBoxOverlaps uses strict < so edge-touching should not collide.
	if def.Collides(a, b) {
		t.Error("edge-touching entities should not collide (strict < check)")
	}
}

// --- Narrow-phase: PreciseCollider ---

// mockPreciseEntity wraps MockEntity and implements PreciseCollider.
// CollidesResult controls the return value of CollidesWith.
type mockPreciseEntity struct {
	*testutil.MockEntity
	CollidesResult bool
	CallCount      int
}

func (m *mockPreciseEntity) CollidesWith(_ def.Entity) bool {
	m.CallCount++
	return m.CollidesResult
}

func TestCollidesCallsPreciseColliderWhenBoundingBoxesOverlap(t *testing.T) {
	a := &mockPreciseEntity{
		MockEntity:     testutil.NewMockEntity(def.EntityTypePlayer),
		CollidesResult: true,
	}
	a.X, a.Y, a.Width, a.Height = 0, 0, 20, 20

	b := testutil.NewMockEntity(def.EntityTypeEnemy)
	b.X, b.Y, b.Width, b.Height = 10, 10, 20, 20

	def.Collides(a, b)

	if a.CallCount == 0 {
		t.Error("CollidesWith was not called on the PreciseCollider")
	}
}

func TestCollidesSkipsPreciseColliderWhenBoundingBoxesMiss(t *testing.T) {
	a := &mockPreciseEntity{
		MockEntity:     testutil.NewMockEntity(def.EntityTypePlayer),
		CollidesResult: true,
	}
	a.X, a.Y, a.Width, a.Height = 0, 0, 10, 10

	b := testutil.NewMockEntity(def.EntityTypeEnemy)
	b.X, b.Y, b.Width, b.Height = 100, 100, 10, 10

	result := def.Collides(a, b)

	if a.CallCount > 0 {
		t.Error("CollidesWith should not be called when bounding boxes miss")
	}
	if result {
		t.Error("Collides = true, want false when bounding boxes miss")
	}
}

func TestCollidesReturnsPreciseColliderResult(t *testing.T) {
	cases := []struct {
		name           string
		collidesResult bool
	}{
		{"precise says yes", true},
		{"precise says no", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := &mockPreciseEntity{
				MockEntity:     testutil.NewMockEntity(def.EntityTypePlayer),
				CollidesResult: tc.collidesResult,
			}
			a.X, a.Y, a.Width, a.Height = 0, 0, 20, 20

			b := testutil.NewMockEntity(def.EntityTypeEnemy)
			b.X, b.Y, b.Width, b.Height = 10, 10, 20, 20

			if got := def.Collides(a, b); got != tc.collidesResult {
				t.Errorf("Collides = %v, want %v", got, tc.collidesResult)
			}
		})
	}
}

func TestCollidesUsesBEntityPreciseColliderWhenAHasNone(t *testing.T) {
	a := testutil.NewMockEntity(def.EntityTypePlayer)
	a.X, a.Y, a.Width, a.Height = 0, 0, 20, 20

	b := &mockPreciseEntity{
		MockEntity:     testutil.NewMockEntity(def.EntityTypeEnemy),
		CollidesResult: false,
	}
	b.X, b.Y, b.Width, b.Height = 10, 10, 20, 20

	result := def.Collides(a, b)

	if b.CallCount == 0 {
		t.Error("CollidesWith was not called on b's PreciseCollider when a has none")
	}
	if result {
		t.Error("Collides = true, want false (precise collider returned false)")
	}
}
