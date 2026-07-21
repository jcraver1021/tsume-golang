package enemy

import (
	"testing"
)

func TestNewMineLoadsSprite(t *testing.T) {
	m, err := NewMine(100, 50)
	if err != nil {
		t.Fatalf("NewMine: %v", err)
	}
	if m == nil {
		t.Fatal("NewMine returned nil")
	}
}

func TestNewRangeMineLoadsSprite(t *testing.T) {
	r, err := NewRangeMine(100, 50)
	if err != nil {
		t.Fatalf("NewRangeMine: %v", err)
	}
	if r == nil {
		t.Fatal("NewRangeMine returned nil")
	}
}

func TestNewPathMineLoadsSprite(t *testing.T) {
	path := []PathSegment{{Frames: 60, VX: 1.0, VY: 0}}
	p, err := NewPathMine(100, 50, path)
	if err != nil {
		t.Fatalf("NewPathMine: %v", err)
	}
	if p == nil {
		t.Fatal("NewPathMine returned nil")
	}
}

func TestNewPathRangeMineLoadsSprite(t *testing.T) {
	path := []PathSegment{{Frames: 60, VX: 1.0, VY: 0}}
	p, err := NewPathRangeMine(100, 50, path)
	if err != nil {
		t.Fatalf("NewPathRangeMine: %v", err)
	}
	if p == nil {
		t.Fatal("NewPathRangeMine returned nil")
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
