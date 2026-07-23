package enemy

import "testing"

func TestMineConstructors(t *testing.T) {
	path := []PathSegment{{Frames: 60, VX: 1.0, VY: 0}}
	testCases := []struct {
		name string
		fn   func() error
	}{
		{"Mine", func() error { _, err := NewMine(100, 50); return err }},
		{"RangeMine", func() error { _, err := NewRangeMine(100, 50); return err }},
		{"PathMine", func() error { _, err := NewPathMine(100, 50, path); return err }},
		{"PathRangeMine", func() error { _, err := NewPathRangeMine(100, 50, path); return err }},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.fn(); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
		})
	}
}

func TestPathRangeMineDetonatesAfterProximityCount(t *testing.T) {
	p, err := NewPathRangeMine(100, 50, nil)
	if err != nil {
		t.Fatal(err)
	}
	if p.ReadyToDetonate() {
		t.Fatal("should not be ready to detonate before any proximity frames")
	}
	p.proximityFrames = pathRangeMineDetonateFrames
	if !p.ReadyToDetonate() {
		t.Fatal("should be ready to detonate after reaching detonateFrames")
	}
}
