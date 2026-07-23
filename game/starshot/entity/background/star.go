package background

import (
	"image/color"
	"math"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
)

// StarVariation defines how a star changes over time
type StarVariation interface {
	// Calculate returns (size, brightness multiplier) based on global tick
	Calculate(tick int) (sizeMultiplier, brightnessMultiplier float64)
}

// NoVariation is a static star with no changes over time
type NoVariation struct{}

func (n *NoVariation) Calculate(tick int) (sizeMultiplier, brightnessMultiplier float64) {
	return 1.0, 1.0
}

// Pulsar oscillates in size and brightness with a sine wave pattern
type Pulsar struct {
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
		frequency:       2 * math.Pi / period,
		sizeAmplitude:   sizeVariation,
		brightAmplitude: brightnessVariation,
	}
}

func (p *Pulsar) Calculate(tick int) (sizeMultiplier, brightnessMultiplier float64) {
	phase := float64(tick) * p.frequency

	wave := math.Sin(phase)

	sizeMultiplier = 1.0 + (p.sizeAmplitude * wave)
	brightnessMultiplier = 1.0 + (p.brightAmplitude * wave)

	return sizeMultiplier, brightnessMultiplier
}

// Twinkle creates random brightness variations (like atmospheric distortion)
type Twinkle struct {
	changeInterval int     // Frames between brightness changes
	variation      float64 // Max variation from 1.0
}

// NewTwinkle creates a twinkling star effect
// changeFrames: how often to pick a new target brightness (e.g., 30 = every 0.5s at 60fps)
// variation: 0.0-1.0, max brightness variation
func NewTwinkle(changeFrames int, variation float64) *Twinkle {
	return &Twinkle{
		changeInterval: changeFrames,
		variation:      variation,
	}
}

func (t *Twinkle) Calculate(tick int) (sizeMultiplier, brightnessMultiplier float64) {
	interval := tick / t.changeInterval

	pseudoRandom := float64((interval*2654435761)%1000) / 1000.0
	targetBrightness := 1.0 + (pseudoRandom-0.5)*2*t.variation

	// Smooth transition within interval (ease in/out)
	progress := float64(tick%t.changeInterval) / float64(t.changeInterval)
	easing := 0.5 - 0.5*math.Cos(progress*math.Pi) // Smooth S-curve

	// Interpolate from 1.0 to target and back
	brightness := 1.0 + (targetBrightness-1.0)*easing

	return 1.0, brightness
}

// Flare creates occasional bright flashes
type Flare struct {
	flareDuration  int
	flareIntensity float64
	minInterval    int
	maxInterval    int
	seed           int // Seed for deterministic randomness
}

// NewFlare creates a star that occasionally flares up
// minFrames, maxFrames: random interval between flares
// duration: how long the flare lasts
// intensity: peak brightness multiplier (e.g., 2.0 = twice as bright)
func NewFlare(minFrames, maxFrames, duration int, intensity float64) *Flare {
	// Use minFrames as seed for this flare's pattern
	return &Flare{
		minInterval:    minFrames,
		maxInterval:    maxFrames,
		flareDuration:  duration,
		flareIntensity: intensity,
		seed:           minFrames, // Unique seed per configuration
	}
}

func (f *Flare) Calculate(tick int) (sizeMultiplier, brightnessMultiplier float64) {
	// Determine flare cycle length
	cycleLength := f.minInterval + f.flareDuration

	// Add variation to cycle start using seed
	offset := (f.seed * 2654435761) % (f.maxInterval - f.minInterval)

	// Where are we in the current cycle?
	adjustedTick := tick + offset
	positionInCycle := adjustedTick % cycleLength

	// Are we in the flare portion of the cycle?
	if positionInCycle >= f.minInterval && positionInCycle < (f.minInterval+f.flareDuration) {
		// Position within flare
		flarePosition := positionInCycle - f.minInterval
		progress := float64(flarePosition) / float64(f.flareDuration)

		// Triangle wave: rise to peak at 50%, fall back
		var brightness float64
		if progress < 0.5 {
			brightness = 1.0 + (f.flareIntensity-1.0)*(progress*2)
		} else {
			brightness = 1.0 + (f.flareIntensity-1.0)*(2-progress*2)
		}

		return 1.0, brightness
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
	currentTick  int // Cache tick for use in Draw()
}

// NewStar creates a static star with no variation
func NewStar(x, y, speed, size int, c color.RGBA) *Star {
	return NewStarWithVariation(x, y, speed, size, c, nil)
}

// NewStarWithVariation creates a star with the given variation behavior
func NewStarWithVariation(x, y, speed, size int, c color.RGBA, variation StarVariation) *Star {
	s := &Star{
		x:           x,
		y:           y,
		speed:       speed,
		baseSize:    size,
		baseColor:   c,
		variation:   variation,
		currentTick: 0,
	}
	s.updateAppearance(0) // Initialize with tick 0
	return s
}

// updateAppearance applies variation effects to compute current appearance
func (s *Star) updateAppearance(tick int) {
	sizeMultiplier := 1.0
	brightnessMultiplier := 1.0

	if s.variation != nil {
		sizeMultiplier, brightnessMultiplier = s.variation.Calculate(tick)
	}

	s.currentSize = max(int(float64(s.baseSize)*sizeMultiplier), 1)

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

func (s *Star) BoundingBoxOverlaps(other def.Entity) bool {
	return false
}

func (s *Star) Act(b def.Scene) {
	// Cache global tick
	s.currentTick = b.Tick()

	s.y += s.speed
	s.updateAppearance(s.currentTick)
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
