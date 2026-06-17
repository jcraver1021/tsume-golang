package background

import (
	"image/color"
	"math"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
)

// ============================================================================
// Star Variation System
// ============================================================================

// StarVariation defines how a star changes over time
type StarVariation interface {
	// Update advances the variation state and returns current (size, brightness multiplier)
	Update() (sizeMultiplier, brightnessMultiplier float64)
}

// NoVariation is a static star with no changes over time
type NoVariation struct{}

func (n *NoVariation) Update() (sizeMultiplier, brightnessMultiplier float64) {
	return 1.0, 1.0
}

// Pulsar oscillates in size and brightness with a sine wave pattern
type Pulsar struct {
	phase           float64 // Current phase in radians
	frequency       float64 // How fast it pulses (radians per frame)
	sizeAmplitude   float64 // How much size varies (0.0 = none, 1.0 = 0-2x)
	brightAmplitude float64 // How much brightness varies
}

// NewPulsar creates a pulsar with the given characteristics
// period: frames per complete cycle (e.g., 60 = 1 second at 60fps)
// sizeVariation: 0.0-1.0, how much size varies (0.5 = varies 50% of base size)
// brightnessVariation: 0.0-1.0, how much brightness varies
func NewPulsar(period float64, sizeVariation, brightnessVariation float64) *Pulsar {
	return &Pulsar{
		phase:           0,
		frequency:       2 * math.Pi / period,
		sizeAmplitude:   sizeVariation,
		brightAmplitude: brightnessVariation,
	}
}

func (p *Pulsar) Update() (sizeMultiplier, brightnessMultiplier float64) {
	// Sine wave oscillation: ranges from -1 to 1
	wave := math.Sin(p.phase)

	// Convert to multipliers: 1.0 +/- amplitude * wave
	sizeMultiplier = 1.0 + (p.sizeAmplitude * wave)
	brightnessMultiplier = 1.0 + (p.brightAmplitude * wave)

	// Advance phase for next frame
	p.phase += p.frequency

	// Keep phase in reasonable range to avoid floating point drift
	if p.phase > 2*math.Pi {
		p.phase -= 2 * math.Pi
	}

	return sizeMultiplier, brightnessMultiplier
}

// Twinkle creates random brightness variations (like atmospheric distortion)
type Twinkle struct {
	frame             int
	changeInterval    int // Frames between brightness changes
	currentBrightness float64
	targetBrightness  float64
	variation         float64 // Max variation from 1.0
}

// NewTwinkle creates a twinkling star effect
// changeFrames: how often to pick a new target brightness (e.g., 30 = every 0.5s at 60fps)
// variation: 0.0-1.0, max brightness variation
func NewTwinkle(changeFrames int, variation float64) *Twinkle {
	return &Twinkle{
		changeInterval:    changeFrames,
		currentBrightness: 1.0,
		targetBrightness:  1.0,
		variation:         variation,
	}
}

func (t *Twinkle) Update() (sizeMultiplier, brightnessMultiplier float64) {
	t.frame++

	// Time to pick a new target?
	if t.frame >= t.changeInterval {
		t.frame = 0
		// Random brightness between (1-variation) and (1+variation)
		pseudoRandom := float64((t.frame*2654435761)%1000) / 1000.0
		t.targetBrightness = 1.0 + (pseudoRandom-0.5)*2*t.variation
	}

	// Smoothly interpolate toward target (easing)
	t.currentBrightness += (t.targetBrightness - t.currentBrightness) * 0.1

	return 1.0, t.currentBrightness
}

// Flare creates occasional bright flashes
type Flare struct {
	frame          int
	nextFlareFrame int
	flareDuration  int
	flareIntensity float64
	flareFrame     int // Current frame within a flare
	minInterval    int
	maxInterval    int
}

// NewFlare creates a star that occasionally flares up
// minFrames, maxFrames: random interval between flares
// duration: how long the flare lasts
// intensity: peak brightness multiplier (e.g., 2.0 = twice as bright)
func NewFlare(minFrames, maxFrames, duration int, intensity float64) *Flare {
	f := &Flare{
		minInterval:    minFrames,
		maxInterval:    maxFrames,
		flareDuration:  duration,
		flareIntensity: intensity,
	}
	f.scheduleNextFlare()
	return f
}

func (f *Flare) scheduleNextFlare() {
	// Simple pseudo-random interval
	range_ := f.maxInterval - f.minInterval
	if range_ <= 0 {
		f.nextFlareFrame = f.frame + f.minInterval
	} else {
		pseudoRandom := (f.frame * 2654435761) % range_
		f.nextFlareFrame = f.frame + f.minInterval + pseudoRandom
	}
}

func (f *Flare) Update() (sizeMultiplier, brightnessMultiplier float64) {
	f.frame++

	// Check if we're in a flare
	if f.flareFrame > 0 {
		// Compute position in flare (0.0 to 1.0 and back)
		progress := float64(f.flareFrame) / float64(f.flareDuration)

		// Triangle wave: rise to peak at 50%, fall back
		var brightness float64
		if progress < 0.5 {
			brightness = 1.0 + (f.flareIntensity-1.0)*(progress*2)
		} else {
			brightness = 1.0 + (f.flareIntensity-1.0)*(2-progress*2)
		}

		f.flareFrame++
		if f.flareFrame > f.flareDuration {
			f.flareFrame = 0
			f.scheduleNextFlare()
		}

		return 1.0, brightness
	}

	// Check if it's time to start a flare
	if f.frame >= f.nextFlareFrame {
		f.flareFrame = 1
	}

	return 1.0, 1.0
}

// ============================================================================
// Star Entity
// ============================================================================

type Star struct {
	x         int
	y         int
	speed     int
	baseSize  int
	baseColor color.RGBA
	variation StarVariation

	// Cached computed values (updated per frame)
	currentSize  int
	currentColor color.RGBA
}

// NewStar creates a static star with no variation
func NewStar(x, y, speed, size int, c color.RGBA) *Star {
	return NewStarWithVariation(x, y, speed, size, c, nil)
}

// NewStarWithVariation creates a star with the given variation behavior
func NewStarWithVariation(x, y, speed, size int, c color.RGBA, variation StarVariation) *Star {
	s := &Star{
		x:         x,
		y:         y,
		speed:     speed,
		baseSize:  size,
		baseColor: c,
		variation: variation,
	}
	s.updateAppearance()
	return s
}

// updateAppearance applies variation effects to compute current appearance
func (s *Star) updateAppearance() {
	sizeMultiplier := 1.0
	brightnessMultiplier := 1.0

	if s.variation != nil {
		sizeMultiplier, brightnessMultiplier = s.variation.Update()
	}

	// Apply size multiplier
	s.currentSize = int(float64(s.baseSize) * sizeMultiplier)
	if s.currentSize < 1 {
		s.currentSize = 1
	}

	// Apply brightness multiplier to color
	s.currentColor = color.RGBA{
		R: uint8(clamp(float64(s.baseColor.R)*brightnessMultiplier, 0, 255)),
		G: uint8(clamp(float64(s.baseColor.G)*brightnessMultiplier, 0, 255)),
		B: uint8(clamp(float64(s.baseColor.B)*brightnessMultiplier, 0, 255)),
		A: s.baseColor.A,
	}
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// SetLocation updates the star's position (used during initial seeding)
func (s *Star) SetLocation(x, y int) {
	s.x = x
	s.y = y
}

func (s *Star) radius() int {
	return s.currentSize >> 1
}

// Entity interface implementation

func (s *Star) Type() def.EntityType {
	return def.EntityTypeBackground
}

func (s *Star) Location() (x, y int) {
	return s.x, s.y
}

func (s *Star) Dimensions() (width, height int) {
	d := s.currentSize
	return d, d
}

func (s *Star) Onscreen(b def.Scene) def.OnScreen {
	r := s.radius()

	if s.y+r < 0 || s.y-r > b.Height() || s.x+r < 0 || s.x-r > b.Width() {
		return def.OffScreen
	}
	if s.y-r > 0 && s.y+r < b.Height() && s.x-r > 0 && s.x+r < b.Width() {
		return def.Fully
	}
	return def.Partially
}

func (s *Star) Overlaps(other def.Entity) bool {
	return false
}

func (s *Star) Act(b def.Scene) {
	s.y += s.speed
	s.updateAppearance()
}

func (s *Star) Draw(img *ebit.Image) {
	r := s.radius()

	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if int(math.Abs(float64(dx)))+int(math.Abs(float64(dy))) <= r {
				alpha := uint8(255 - (int(math.Abs(float64(dx)))+int(math.Abs(float64(dy))))*255/(r+1))
				img.Set(s.x+dx, s.y+dy, color.RGBA{
					R: s.currentColor.R,
					G: s.currentColor.G,
					B: s.currentColor.B,
					A: alpha,
				})
			}
		}
	}
}

func (s *Star) CanBeRemoved() bool {
	return s.y-s.radius() > def.ScreenHeight
}
