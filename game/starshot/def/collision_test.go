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

	// you gotta keep 'em separated
	return a, b
}

func TestCollides(t *testing.T) {
	testCases := []struct {
		name      string
		generator func() (*testutil.MockEntity, *testutil.MockEntity)
		want      bool
	}{
		{
			name:      "overlapping",
			generator: overlapping,
			want:      true,
		},
		{
			name:      "separated",
			generator: separated,
			want:      false,
		},
		{
			name: "touching edge",
			generator: func() (*testutil.MockEntity, *testutil.MockEntity) {
				a := testutil.NewMockEntity(def.EntityTypePlayer)
				a.X, a.Y, a.Width, a.Height = 0, 0, 10, 10

				b := testutil.NewMockEntity(def.EntityTypeEnemy)
				b.X, b.Y, b.Width, b.Height = 10, 0, 10, 10

				return a, b
			},
			want: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a, b := tc.generator()
			if got := def.Collides(a, b); got != tc.want {
				t.Errorf("Collides = %v, want %v", got, tc.want)
			}
			if got := def.Collides(b, a); got != tc.want {
				t.Errorf("Collides = %v, want %v", got, tc.want)
			}
		})
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
	// CollidesWith should not be called if the bounding boxes do not overlap
	// We increment to track collisions between precise and non-precise colliders.
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
	testCases := []struct {
		name string
		want bool
	}{
		{
			name: "precise says yes",
			want: true,
		},
		{
			name: "precise says no",
			want: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := &mockPreciseEntity{
				MockEntity:     testutil.NewMockEntity(def.EntityTypePlayer),
				CollidesResult: tc.want,
			}
			a.X, a.Y, a.Width, a.Height = 0, 0, 20, 20

			b := testutil.NewMockEntity(def.EntityTypeEnemy)
			b.X, b.Y, b.Width, b.Height = 10, 10, 20, 20

			if got := def.Collides(a, b); got != tc.want {
				t.Errorf("Collides = %v, want %v", got, tc.want)
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
