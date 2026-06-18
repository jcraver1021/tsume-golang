package background_test

import (
	"image/color"
	"testing"

	"tsumegolang/game/starshot/def"
	. "tsumegolang/game/starshot/entity/background"
	"tsumegolang/game/starshot/testutil"
)

// ============================================================================
// Banner Entity Tests
// ============================================================================

func TestBannerType(t *testing.T) {
	banner, err := NewBanner("Test", 100, 100, 24.0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		t.Fatalf("NewBanner failed: %v", err)
	}

	got := banner.Type()
	want := def.EntityTypeBackground

	if got != want {
		t.Errorf("Type() = %v, want %v", got, want)
	}
}

func TestBannerLocation(t *testing.T) {
	banner, err := NewBanner("Test", 100, 200, 24.0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		t.Fatalf("NewBanner failed: %v", err)
	}

	gotX, gotY := banner.Location()
	wantX, wantY := 100, 200

	if gotX != wantX || gotY != wantY {
		t.Errorf("Location() = (%d, %d), want (%d, %d)", gotX, gotY, wantX, wantY)
	}
}

func TestBannerDimensions(t *testing.T) {
	banner, err := NewBanner("Test", 0, 0, 24.0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		t.Fatalf("NewBanner failed: %v", err)
	}

	gotW, gotH := banner.Dimensions()

	// Dimensions should be positive (text has some width/height)
	if gotW <= 0 || gotH <= 0 {
		t.Errorf("Dimensions() = (%d, %d), want positive values", gotW, gotH)
	}
}

func TestBannerOverlaps(t *testing.T) {
	banner, err := NewBanner("Test", 100, 100, 24.0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		t.Fatalf("NewBanner failed: %v", err)
	}

	// Banners should not overlap (non-interactive background elements)
	if banner.BoundingBoxOverlaps(banner) {
		t.Error("banners should not overlap")
	}
}

func TestBannerPermanent(t *testing.T) {
	banner, err := NewBanner("Test", 100, 100, 24.0, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		t.Fatalf("NewBanner failed: %v", err)
	}

	// Permanent banners should never be removed
	if banner.CanBeRemoved() {
		t.Error("permanent banner should never be removed")
	}

	// Even after many Act() calls
	scene := testutil.NewMockScene()
	for range 1000 {
		banner.Act(scene)
	}

	if banner.CanBeRemoved() {
		t.Error("permanent banner should never be removed after Act calls")
	}
}

func TestBannerTemporary(t *testing.T) {
	opts := BannerOptions{
		Text:           "Temporary",
		X:              100,
		Y:              100,
		FontSize:       24.0,
		TextColor:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
		DurationFrames: 60, // 1 second at 60fps
	}

	banner, err := NewBannerWithOptions(opts)
	if err != nil {
		t.Fatalf("NewBannerWithOptions failed: %v", err)
	}

	// Should not be removed initially
	if banner.CanBeRemoved() {
		t.Error("temporary banner should not be removed before duration")
	}

	scene := testutil.NewMockScene()

	// Act for duration - 1 frames
	for range 59 {
		banner.Act(scene)
	}

	// Still should not be removed
	if banner.CanBeRemoved() {
		t.Error("temporary banner should not be removed before duration")
	}

	// One more Act should trigger removal
	banner.Act(scene)

	if !banner.CanBeRemoved() {
		t.Error("temporary banner should be removed after duration")
	}
}

func TestBannerActIncrementsFrames(t *testing.T) {
	opts := BannerOptions{
		Text:           "Test",
		X:              100,
		Y:              100,
		FontSize:       24.0,
		TextColor:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
		DurationFrames: 10,
	}

	banner, err := NewBannerWithOptions(opts)
	if err != nil {
		t.Fatalf("NewBannerWithOptions failed: %v", err)
	}

	scene := testutil.NewMockScene()

	// Act exactly duration times
	for range 10 {
		banner.Act(scene)
	}

	// Should be removable now
	if !banner.CanBeRemoved() {
		t.Error("banner should be removable after exactly duration frames")
	}
}

func TestBannerWithBackgroundColor(t *testing.T) {
	bgColor := color.RGBA{R: 0, G: 0, B: 0, A: 128}
	opts := BannerOptions{
		Text:            "Test",
		X:               100,
		Y:               100,
		FontSize:        24.0,
		TextColor:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
		BackgroundColor: &bgColor,
		Padding:         10,
	}

	banner, err := NewBannerWithOptions(opts)
	if err != nil {
		t.Fatalf("NewBannerWithOptions failed: %v", err)
	}

	// Should create without error (visual verification would require graphics context)
	if banner == nil {
		t.Error("banner with background color should not be nil")
	}
}
