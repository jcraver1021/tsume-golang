package environment

import (
	"image/color"
	"math/rand"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/background"

	ebit "github.com/hajimehoshi/ebiten/v2"
)

type Space struct {
	// parameters for procedural generation
	starDensity float64
	// TODO: add nebulaDensity and planetDensity when we implement those features
}

type Layer int

const (
	Close Layer = iota
	Mid
	Far
)

// String returns the string representation of a Layer for testing/debugging
func (l Layer) String() string {
	switch l {
	case Close:
		return "Close"
	case Mid:
		return "Mid"
	case Far:
		return "Far"
	default:
		return "Unknown"
	}
}

// Stars - Layer-based generation with subtle color variance

// SizeAndSpeedForLayer returns the size and speed for a given parallax layer.
// Exported for testing the parallax contracts.
func SizeAndSpeedForLayer(layer Layer) (size, speed int) {
	switch layer {
	case Close:
		return 5, 5
	case Mid:
		return 3, 3
	case Far:
		return 2, 1
	}
	return 1, 1
}

// generateStarColor creates layer-appropriate colors with subtle variance
func generateStarColor(layer Layer) color.RGBA {
	switch layer {
	case Close:
		return generateCloseLayerColor()
	case Mid:
		return generateMidLayerColor()
	case Far:
		return generateFarLayerColor()
	}
	return color.RGBA{R: 255, G: 255, B: 255, A: 255}
}

// Close layer: Bright whites with occasional warm tints (orange/yellow)
func generateCloseLayerColor() color.RGBA {
	roll := rand.Float64()

	if roll < 0.70 {
		// 70% - Bright white stars
		brightness := uint8(220 + rand.Intn(36)) // 220-255
		return color.RGBA{R: brightness, G: brightness, B: brightness, A: 255}
	} else if roll < 0.85 {
		// 15% - Warm yellow/orange tint
		base := uint8(240 + rand.Intn(16)) // 240-255
		return color.RGBA{R: base, G: base, B: uint8(180 + rand.Intn(40)), A: 255}
	} else {
		// 15% - Warm orange tint
		return color.RGBA{R: 255, G: uint8(200 + rand.Intn(56)), B: uint8(150 + rand.Intn(51)), A: 255}
	}
}

// Mid layer: Mix of white and subtle colored stars
func generateMidLayerColor() color.RGBA {
	roll := rand.Float64()

	if roll < 0.50 {
		// 50% - White/gray stars
		brightness := uint8(180 + rand.Intn(76)) // 180-255
		return color.RGBA{R: brightness, G: brightness, B: brightness, A: 255}
	} else if roll < 0.70 {
		// 20% - Pale yellow
		base := uint8(200 + rand.Intn(56))
		return color.RGBA{R: base, G: base, B: uint8(170 + rand.Intn(51)), A: 255}
	} else if roll < 0.85 {
		// 15% - Pale cyan
		base := uint8(180 + rand.Intn(51))
		return color.RGBA{R: base, G: uint8(200 + rand.Intn(56)), B: uint8(220 + rand.Intn(36)), A: 255}
	} else {
		// 15% - Pale pink/magenta
		base := uint8(200 + rand.Intn(56))
		return color.RGBA{R: base, G: uint8(180 + rand.Intn(51)), B: base, A: 255}
	}
}

// Far layer: Cool-tinted stars (blue/purple) for atmospheric depth
func generateFarLayerColor() color.RGBA {
	roll := rand.Float64()

	if roll < 0.40 {
		// 40% - Cool white/gray (subtle blue tint)
		brightness := uint8(150 + rand.Intn(51)) // 150-200
		blueTint := uint8(int(brightness) + 20 + rand.Intn(31))
		return color.RGBA{R: brightness, G: brightness, B: blueTint, A: 255}
	} else if roll < 0.70 {
		// 30% - Blue stars
		base := uint8(140 + rand.Intn(41))
		return color.RGBA{R: base, G: base, B: uint8(170 + rand.Intn(51)), A: 255}
	} else if roll < 0.90 {
		// 20% - Purple/violet stars
		base := uint8(150 + rand.Intn(41))
		return color.RGBA{R: uint8(160 + rand.Intn(41)), G: base, B: uint8(180 + rand.Intn(51)), A: 255}
	} else {
		// 10% - Cyan stars
		base := uint8(140 + rand.Intn(41))
		return color.RGBA{R: base, G: uint8(170 + rand.Intn(51)), B: uint8(180 + rand.Intn(51)), A: 255}
	}
}

// newStar creates a star with layer-appropriate properties
func newStar(x, y int, layer Layer) *background.Star {
	size, speed := SizeAndSpeedForLayer(layer)
	c := generateStarColor(layer)

	// Add variation to some stars
	variation := generateVariation(layer)

	return background.NewStarWithVariation(x, y, speed, size, c, variation)
}

// generateVariation creates occasional special star effects
func generateVariation(layer Layer) background.StarVariation {
	roll := rand.Float64()

	// Only add variations to larger stars (Mid and Close layers)
	if layer == Far {
		return nil // Far stars are too small for visible variation
	}

	// 10% of stars get special effects
	if roll < 0.10 {
		effectRoll := rand.Float64()

		if effectRoll < 0.6 {
			// 60% pulsars - slow, rhythmic pulse
			period := 60.0 + rand.Float64()*120.0 // 1-3 seconds at 60fps
			sizeVar := 0.3 + rand.Float64()*0.2   // 30-50% size variation
			brightVar := 0.4 + rand.Float64()*0.3 // 40-70% brightness variation
			return background.NewPulsar(period, sizeVar, brightVar)

		} else if effectRoll < 0.9 {
			// 30% twinkle - random atmospheric shimmer
			changeInterval := 15 + rand.Intn(30)  // Change every 0.25-0.75s
			variation := 0.2 + rand.Float64()*0.3 // 20-50% brightness variation
			return background.NewTwinkle(changeInterval, variation)

		} else {
			// 10% flare - occasional bright flash
			minInterval := 180 + rand.Intn(240)   // 3-7 seconds between flares
			maxInterval := minInterval + 180      // +3 seconds
			duration := 30 + rand.Intn(30)        // 0.5-1 second duration
			intensity := 1.5 + rand.Float64()*1.0 // 1.5-2.5x brightness
			return background.NewFlare(minInterval, maxInterval, duration, intensity)
		}
	}

	return nil // 90% are static stars
}

func NewSpace(starDensity float64, b def.Scene) *Space {
	s := &Space{
		starDensity: starDensity,
	}
	s.seedInitialStars(b)
	return s
}

func (s *Space) seedInitialStars(b def.Scene) {
	for y := range b.Height() {
		if star := s.maybeAddStar(Close, b); star != nil {
			x, _ := star.Location()
			star.SetLocation(x, y)
			b.Entities().Add(star)
		}
		if star := s.maybeAddStar(Mid, b); star != nil {
			x, _ := star.Location()
			star.SetLocation(x, y)
			b.Entities().Add(star)
		}
		if star := s.maybeAddStar(Far, b); star != nil {
			x, _ := star.Location()
			star.SetLocation(x, y)
			b.Entities().Add(star)
		}
	}
}

func (s *Space) Type() def.EntityType {
	return def.EntityTypeEnvironment
}

func (s *Space) Onscreen(_ def.Scene) def.OnScreen {
	return def.OffScreen
}

func (s *Space) Location() (x, y int) {
	return 0, 0
}

func (s *Space) Dimensions() (width, height int) {
	return def.ScreenWidth, def.ScreenHeight
}

func (s *Space) BoundingBoxOverlaps(other def.Entity) bool {
	return false
}

func (s *Space) maybeAddStar(layer Layer, b def.Scene) *background.Star {
	if rand.Float64() < s.starDensity {
		x := rand.Intn(b.Width())
		y := -10 // start above the screen
		return newStar(x, y, layer)
	}

	return nil
}

func (s *Space) Act(b def.Scene) {
	if star := s.maybeAddStar(Close, b); star != nil {
		b.Entities().Add(star)
	}
	if star := s.maybeAddStar(Mid, b); star != nil {
		b.Entities().Add(star)
	}
	if star := s.maybeAddStar(Far, b); star != nil {
		b.Entities().Add(star)
	}
}

func (s *Space) Draw(img *ebit.Image) {}

func (s *Space) CanBeRemoved() bool {
	return false
}
