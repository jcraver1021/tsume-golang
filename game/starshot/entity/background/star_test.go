package background_test

import (
	"image/color"
	"testing"

	"tsumegolang/game/starshot/def"
	. "tsumegolang/game/starshot/entity/background"
	"tsumegolang/game/starshot/testutil"
)

// ============================================================================
// Star Entity Tests
// ============================================================================

func TestStarLocation(t *testing.T) {
	star := NewStar(10, 20, 1, 2, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	gotX, gotY := star.Location()
	if gotX != 10 || gotY != 20 {
		t.Errorf("Location() = (%d, %d), want (10, 20)", gotX, gotY)
	}

	star.SetLocation(30, 40)
	gotX, gotY = star.Location()
	if gotX != 30 || gotY != 40 {
		t.Errorf("Location() after SetLocation = (%d, %d), want (30, 40)", gotX, gotY)
	}
}

func TestStarDimensions(t *testing.T) {
	star := NewStar(0, 0, 1, 4, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	gotW, gotH := star.Dimensions()
	if gotW != 4 || gotH != 4 {
		t.Errorf("Dimensions() = (%d, %d), want (4, 4)", gotW, gotH)
	}
}

func TestStarOnscreen(t *testing.T) {
	scene := testutil.NewMockSceneWithSize(100, 100)

	testCases := []struct {
		name string
		x    int
		y    int
		want def.OnScreen
	}{
		{"fully on-screen", 50, 50, def.Fully},
		{"partially on-screen", 98, 50, def.Partially},
		{"off-screen", 150, 150, def.OffScreen},
	}

	star := NewStar(0, 0, 1, 4, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			star.SetLocation(tc.x, tc.y)
			got := star.Onscreen(scene)
			if got != tc.want {
				t.Errorf("Onscreen() at (%d,%d) = %v, want %v", tc.x, tc.y, got, tc.want)
			}
		})
	}
}

func TestStarOverlaps(t *testing.T) {
	star := NewStar(10, 10, 1, 4, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	// Stars should not overlap (gameplay design choice); all background elements are non-interactive
	if star.Overlaps(star) {
		t.Error("stars should not overlap")
	}
}

func TestStarAct(t *testing.T) {
	star := NewStar(50, 50, 2, 4, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	star.Act(nil) // Scene not used in Act

	gotX, gotY := star.Location()
	wantX, wantY := 50, 52
	if gotX != wantX || gotY != wantY {
		t.Errorf("Location() after Act = (%d, %d), want (%d, %d)", gotX, gotY, wantX, wantY)
	}
}

// ============================================================================
// Variation Tests
// ============================================================================

func TestNoVariation(t *testing.T) {
	v := &NoVariation{}

	for range 100 {
		gotSize, gotBright := v.Update()

		if gotSize != 1.0 {
			t.Errorf("NoVariation size = %v, want 1.0", gotSize)
		}

		if gotBright != 1.0 {
			t.Errorf("NoVariation brightness = %v, want 1.0", gotBright)
		}
	}
}

func TestPulsarOscillates(t *testing.T) {
	period := 60.0
	sizeVar := 0.5
	brightVar := 0.5

	pulsar := NewPulsar(period, sizeVar, brightVar)

	var minSize, maxSize float64 = 2.0, 0.0
	var minBright, maxBright float64 = 2.0, 0.0

	// Run for one complete period
	for range 60 {
		size, bright := pulsar.Update()

		if size < minSize {
			minSize = size
		}
		if size > maxSize {
			maxSize = size
		}
		if bright < minBright {
			minBright = bright
		}
		if bright > maxBright {
			maxBright = bright
		}
	}

	// Should oscillate around 1.0
	// With 0.5 variation: should reach roughly 0.5 and 1.5
	wantMinSize := 0.5 // Exact: 1.0 - 0.5 = 0.5
	wantMaxSize := 1.5 // Exact: 1.0 + 0.5 = 1.5

	tolerance := 0.1

	if minSize > wantMinSize+tolerance {
		t.Errorf("Pulsar min size = %v, want ~%v", minSize, wantMinSize)
	}

	if maxSize < wantMaxSize-tolerance {
		t.Errorf("Pulsar max size = %v, want ~%v", maxSize, wantMaxSize)
	}

	// Brightness should behave similarly
	if minBright > wantMinSize+tolerance {
		t.Errorf("Pulsar min brightness = %v, want ~%v", minBright, wantMinSize)
	}

	if maxBright < wantMaxSize-tolerance {
		t.Errorf("Pulsar max brightness = %v, want ~%v", maxBright, wantMaxSize)
	}
}

func TestPulsarPeriodic(t *testing.T) {
	period := 60.0
	pulsar := NewPulsar(period, 0.5, 0.5)

	// Get initial value
	initialSize, initialBright := pulsar.Update()

	// Advance one full period (minus the one we just consumed)
	for range 59 {
		pulsar.Update()
	}

	// After one period, should be back to start
	gotSize, gotBright := pulsar.Update()

	tolerance := 0.01

	if abs(gotSize-initialSize) > tolerance {
		t.Errorf("after one period, size = %v, want ~%v", gotSize, initialSize)
	}

	if abs(gotBright-initialBright) > tolerance {
		t.Errorf("after one period, brightness = %v, want ~%v", gotBright, initialBright)
	}
}

func TestTwinkleWithinBounds(t *testing.T) {
	variation := 0.3
	twinkle := NewTwinkle(10, variation)

	wantMin := 1.0 - variation - 0.1 // Small tolerance
	wantMax := 1.0 + variation + 0.1

	for range 200 {
		_, bright := twinkle.Update()

		if bright < wantMin || bright > wantMax {
			t.Errorf("Twinkle brightness = %v, want in range [%v, %v]", bright, wantMin, wantMax)
		}
	}
}

func TestTwinkleVaries(t *testing.T) {
	twinkle := NewTwinkle(10, 0.5)

	values := make([]float64, 100)
	for i := range values {
		_, bright := twinkle.Update()
		values[i] = bright
	}

	// Check that we have some variation (not all the same)
	allSame := true
	first := values[0]
	for _, v := range values[1:] {
		if abs(v-first) > 0.01 {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Twinkle should vary over time, got constant brightness")
	}
}

func TestFlarePattern(t *testing.T) {
	minInterval := 50
	maxInterval := 100
	duration := 10
	intensity := 2.0

	flare := NewFlare(minInterval, maxInterval, duration, intensity)

	normalCount := 0
	elevatedCount := 0

	// Run for enough frames to see at least one flare
	for range 150 {
		_, bright := flare.Update()

		if bright > 1.1 {
			elevatedCount++
		} else {
			normalCount++
		}
	}

	// Should be mostly normal, with some elevated
	if elevatedCount == 0 {
		t.Error("Flare should have some elevated brightness, got none")
	}

	if normalCount < elevatedCount {
		t.Errorf("Flare should be mostly normal brightness, got normal=%d elevated=%d", normalCount, elevatedCount)
	}
}

// ============================================================================
// Integration Tests (skipped - require graphics context)
// ============================================================================

func TestStarWithVariation(t *testing.T) {
	t.Skip("Skipping star integration test - requires graphics context")

	pulsar := NewPulsar(60, 0.5, 0.5)
	star := NewStarWithVariation(100, 100, 1, 10, testColor(), pulsar)

	// Act should update the variation
	star.Act(nil)

	// Dimensions should reflect varied size
	w, h := star.Dimensions()

	// With 0.5 variation on size 10, should be between 5 and 15
	wantMin := 5
	wantMax := 15

	if w < wantMin || w > wantMax {
		t.Errorf("Star with pulsar dimensions = (%d, %d), want width in [%d, %d]", w, h, wantMin, wantMax)
	}
}

func TestStarWithoutVariation(t *testing.T) {
	t.Skip("Skipping star integration test - requires graphics context")

	star := NewStar(100, 100, 1, 10, testColor())

	w1, h1 := star.Dimensions()

	// Act multiple times
	for range 100 {
		star.Act(nil)
	}

	w2, h2 := star.Dimensions()

	if w1 != w2 || h1 != h2 {
		t.Errorf("Static star dimensions changed from (%d,%d) to (%d,%d), want constant", w1, h1, w2, h2)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func testColor() color.RGBA {
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}
