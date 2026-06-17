package player

import "image/color"

// SpriteComponent represents a drawable component of the player ship
type SpriteComponent struct {
	Name   string
	Width  int
	Height int
	// Sprite data: each row is a slice of ColorCode values
	// Coordinates are relative to component origin (0,0 = top-left)
	Data [][]ColorCode
	// Offset from player's origin (x, y)
	OffsetX int
	OffsetY int
}

// ColorCode represents a pixel color, can be static or animated
type ColorCode int

const (
	ColorEmpty ColorCode = iota
	ColorWhite
	ColorLightGray
	ColorDarkGray
	ColorLightBlue
	ColorEngineGlow1 // Brightest
	ColorEngineGlow2 // Medium
	ColorEngineGlow3 // Dim
	ColorRedAccent
	ColorOrangeAccent
)

// AnimationSequence defines how a pixel changes over time
type AnimationSequence struct {
	Frames        []ColorCode // Color sequence
	FrameDuration int         // Frames to show each color
}

// GetColorAtFrame returns the color for the given animation frame
func (a *AnimationSequence) GetColorAtFrame(frame int) ColorCode {
	if len(a.Frames) == 0 {
		return ColorEmpty
	}
	position := (frame / a.FrameDuration) % len(a.Frames)
	return a.Frames[position]
}

// AnimatedPixel represents a pixel that changes color over time
type AnimatedPixel struct {
	X        int
	Y        int
	Sequence AnimationSequence
}

// GetColorPalette returns the color palette for all color codes
func GetColorPalette() map[ColorCode]color.RGBA {
	return map[ColorCode]color.RGBA{
		ColorEmpty:        {R: 0, G: 0, B: 0, A: 0},
		ColorWhite:        {R: 240, G: 240, B: 240, A: 255},
		ColorLightGray:    {R: 180, G: 180, B: 180, A: 255},
		ColorDarkGray:     {R: 100, G: 100, B: 110, A: 255},
		ColorLightBlue:    {R: 100, G: 180, B: 255, A: 255},
		ColorEngineGlow1:  {R: 255, G: 220, B: 120, A: 255}, // Brightest
		ColorEngineGlow2:  {R: 255, G: 200, B: 100, A: 255}, // Medium
		ColorEngineGlow3:  {R: 220, G: 160, B: 80, A: 255},  // Dim
		ColorRedAccent:    {R: 255, G: 100, B: 100, A: 255},
		ColorOrangeAccent: {R: 255, G: 150, B: 80, A: 255},
	}
}

// Standard engine pulse animation
func EngineGlowAnimation() AnimationSequence {
	return AnimationSequence{
		Frames:        []ColorCode{ColorEngineGlow1, ColorEngineGlow2, ColorEngineGlow3, ColorEngineGlow2},
		FrameDuration: 4, // 4 game frames per animation frame
	}
}

// Define standard ship components

// CoreHull returns the main body of the ship
func CoreHull() *SpriteComponent {
	return &SpriteComponent{
		Name:    "core_hull",
		Width:   32,
		Height:  32,
		OffsetX: 0,
		OffsetY: 0,
		Data: [][]ColorCode{
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // Nose
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 4, 4, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // Cockpit
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 4, 4, 4, 4, 4, 4, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 4, 4, 4, 4, 4, 4, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 1, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 1, 1, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 1, 1, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 1, 1, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 1, 1, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 1, 1, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 1, 1, 0, 0, 0, 0},
			{0, 0, 0, 1, 1, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 1, 1, 0, 0, 0},
			{0, 0, 1, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 1, 0, 0},
			{0, 1, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 1, 0},
			{1, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 3, 1},
			{0, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 3, 0},
			{0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 1, 1, 1, 1, 1, 1, 3, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 1, 1, 1, 1, 3, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 1, 1, 3, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 2, 1, 1, 3, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // Engine space
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
}

// BasicEngine returns a simple engine component
func BasicEngine() *SpriteComponent {
	return &SpriteComponent{
		Name:    "basic_engine",
		Width:   32,
		Height:  3,
		OffsetX: 0,
		OffsetY: 29, // Bottom of ship
		Data: [][]ColorCode{
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 5, 5, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 5, 5, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
}

// BasicEngineAnimatedPixels returns animated pixels for basic engine glow
func BasicEngineAnimatedPixels() []AnimatedPixel {
	anim := EngineGlowAnimation()
	return []AnimatedPixel{
		{X: 15, Y: 29, Sequence: anim},
		{X: 16, Y: 29, Sequence: anim},
		{X: 14, Y: 30, Sequence: anim},
		{X: 15, Y: 30, Sequence: anim},
		{X: 16, Y: 30, Sequence: anim},
		{X: 17, Y: 30, Sequence: anim},
		{X: 15, Y: 31, Sequence: anim},
		{X: 16, Y: 31, Sequence: anim},
	}
}

// CentralCannon returns a nose-mounted cannon component
func CentralCannon() *SpriteComponent {
	return &SpriteComponent{
		Name:    "central_cannon",
		Width:   4,
		Height:  2,
		OffsetX: 14,
		OffsetY: 0,
		Data: [][]ColorCode{
			{0, 9, 9, 0},
			{9, 9, 9, 9},
		},
	}
}

// WingGuns returns wing-mounted gun pods
func WingGuns() *SpriteComponent {
	return &SpriteComponent{
		Name:    "wing_guns",
		Width:   32,
		Height:  4,
		OffsetX: 0,
		OffsetY: 14, // At widest wingspan
		Data: [][]ColorCode{
			{8, 8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 8},
			{9, 8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 9},
			{9, 9, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 9},
			{8, 9, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 8},
		},
	}
}
