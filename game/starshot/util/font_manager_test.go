package util_test

import (
	"testing"

	. "tsumegolang/game/starshot/util"
)

func TestFontManagerCreation(t *testing.T) {
	fm, err := NewFontManager()
	if err != nil {
		t.Fatalf("NewFontManager() failed: %v", err)
	}

	if fm == nil {
		t.Error("NewFontManager() returned nil manager")
	}
}

func TestGetFace(t *testing.T) {
	fm, err := NewFontManager()
	if err != nil {
		t.Fatalf("NewFontManager() failed: %v", err)
	}

	sizes := []float64{12.0, 24.0, 48.0}

	for _, size := range sizes {
		face := fm.GetFace(size)
		if face == nil {
			t.Errorf("GetFace(%v) returned nil", size)
		}

		if face.Size != size {
			t.Errorf("GetFace(%v) face size = %v, want %v", size, face.Size, size)
		}
	}
}

func TestGetFaceCaching(t *testing.T) {
	fm, err := NewFontManager()
	if err != nil {
		t.Fatalf("NewFontManager() failed: %v", err)
	}

	size := 24.0

	// Get face twice
	face1 := fm.GetFace(size)
	face2 := fm.GetFace(size)

	// Should return same instance (cached)
	if face1 != face2 {
		t.Error("GetFace should return cached face for same size")
	}
}

func TestGetDefaultFontManager(t *testing.T) {
	fm1, err1 := GetDefaultFontManager()
	if err1 != nil {
		t.Fatalf("GetDefaultFontManager() failed: %v", err1)
	}

	fm2, err2 := GetDefaultFontManager()
	if err2 != nil {
		t.Fatalf("GetDefaultFontManager() second call failed: %v", err2)
	}

	// Should return same singleton instance
	if fm1 != fm2 {
		t.Error("GetDefaultFontManager should return same singleton instance")
	}
}
