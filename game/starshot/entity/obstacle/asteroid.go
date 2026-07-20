package obstacle

import (
	"image/color"
	"math/rand"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
)

type AsteroidSize int

const (
	AsteroidTiny AsteroidSize = iota
	AsteroidSmall
	AsteroidMedium
	AsteroidLarge
	AsteroidHuge
	AsteroidMassive
	AsteroidGigantic
	AsteroidColossal
)

func (s AsteroidSize) Dimensions() (width, height int) {
	switch s {
	case AsteroidTiny:
		return 8, 8
	case AsteroidSmall:
		return 12, 12
	case AsteroidMedium:
		return 20, 20
	case AsteroidLarge:
		return 32, 32
	case AsteroidHuge:
		return 48, 48
	case AsteroidMassive:
		return 64, 64
	case AsteroidGigantic:
		return 80, 80
	case AsteroidColossal:
		return 96, 96
	default:
		return 0, 0
	}
}

func (s AsteroidSize) Speed() int {
	switch s {
	case AsteroidTiny:
		return 4
	case AsteroidSmall:
		return 3
	case AsteroidMedium:
		return 2
	case AsteroidLarge:
		return 2
	case AsteroidHuge:
		return 1
	case AsteroidMassive:
		return 1
	case AsteroidGigantic:
		return 1
	case AsteroidColossal:
		return 1
	default:
		return 1
	}
}

func (s AsteroidSize) HP() int {
	switch s {
	case AsteroidTiny:
		return 1
	case AsteroidSmall:
		return 2
	case AsteroidMedium:
		return 3
	case AsteroidLarge:
		return 5
	case AsteroidHuge:
		return 8
	case AsteroidMassive:
		return 12
	case AsteroidGigantic:
		return 18
	case AsteroidColossal:
		return 25
	default:
		return 1
	}
}

// Mass scales inversely with impulse sensitivity; larger rocks barely move.
func (s AsteroidSize) Mass() float64 {
	switch s {
	case AsteroidTiny:
		return 1
	case AsteroidSmall:
		return 2
	case AsteroidMedium:
		return 4
	case AsteroidLarge:
		return 8
	case AsteroidHuge:
		return 16
	case AsteroidMassive:
		return 32
	case AsteroidGigantic:
		return 48
	case AsteroidColossal:
		return 64
	default:
		return 1
	}
}

// Asteroid is a ColorMatrix-based asteroid with procedural multi-color generation
type Asteroid struct {
	x, y          int
	fx, fy        float64 // sub-pixel position
	width, height int
	vx, vy        float64 // velocity in pixels/frame
	size          AsteroidSize
	sprite        *draw.ColorMatrix

	hp    int
	maxHP int
	dead  bool
}

// NewAsteroid creates a new procedurally-generated multi-colored asteroid
func NewAsteroid(x, y int, size AsteroidSize) *Asteroid {
	width, height := size.Dimensions()
	hp := size.HP()

	// Generate procedural asteroid sprite
	sprite := generateAsteroidSprite(width, height, size)

	return &Asteroid{
		x:      x,
		y:      y,
		fx:     float64(x),
		fy:     float64(y),
		width:  width,
		height: height,
		vx:     0,
		vy:     float64(size.Speed()),
		size:   size,
		sprite: sprite,
		hp:     hp,
		maxHP:  hp,
	}
}

// newSplitAsteroid creates a child asteroid with an initial velocity offset, used when splitting.
func newSplitAsteroid(x, y int, size AsteroidSize, vx, vy float64) *Asteroid {
	a := NewAsteroid(x, y, size)
	a.vx = vx
	a.vy = vy
	return a
}

// NewRandomAsteroid creates an asteroid with random size from a given range
func NewRandomAsteroid(x, y int) *Asteroid {
	// Default distribution (for backwards compatibility)
	return NewRandomAsteroidInRange(x, y, AsteroidSmall, AsteroidLarge)
}

// NewRandomAsteroidInRange creates a random asteroid within a size range
func NewRandomAsteroidInRange(x, y int, minSize, maxSize AsteroidSize) *Asteroid {
	// Random size within the inclusive range
	sizeRange := int(maxSize - minSize + 1)
	size := AsteroidSize(int(minSize) + rand.Intn(sizeRange))
	return NewAsteroid(x, y, size)
}

func (a *Asteroid) Type() def.EntityType {
	return def.EntityTypeObstacle
}

func (a *Asteroid) Location() (x, y int) {
	return a.x, a.y
}

func (a *Asteroid) Dimensions() (width, height int) {
	return a.width, a.height
}

func (a *Asteroid) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(a.x+a.width < ox || a.x > ox+ow || a.y+a.height < oy || a.y > oy+oh)
}

func (a *Asteroid) Act(b def.Scene) {
	if a.dead {
		return
	}
	a.fx += a.vx
	a.fy += a.vy
	// Dampen horizontal drift so asteroids don't fly off forever
	a.vx *= 0.98
	a.x = int(a.fx)
	a.y = int(a.fy)
}

func (a *Asteroid) Draw(img *ebit.Image) {
	pixels := a.sprite.Render()

	for row := range pixels {
		for col := range pixels[row] {
			c := pixels[row][col]
			if c.A > 0 { // Only draw non-transparent pixels
				img.Set(a.x+col, a.y+row, c)
			}
		}
	}
}

func (a *Asteroid) CanBeRemoved() bool {
	return a.dead || a.y > def.ScreenHeight
}

// --- Mortal ---

func (a *Asteroid) GetDeathEffect() def.DeathEffect {
	size := def.ExplosionSmall
	if a.size >= AsteroidMedium {
		size = def.ExplosionMedium
	}
	return def.DeathEffect{
		ExplosionSize:    size,
		SlowdownDuration: 0,
	}
}

func (a *Asteroid) MarkAsDead(scene def.Scene) {
	a.dead = true
	// Spawn two smaller asteroids if there is a next size down
	if a.size <= AsteroidTiny {
		return
	}
	childSize := a.size - 1
	cw, _ := childSize.Dimensions()
	cx := a.x + a.width/2
	cy := a.y + a.height/4
	spread := 1.5

	child1 := newSplitAsteroid(cx-cw-2, cy, childSize, a.vx-spread, a.vy)
	child2 := newSplitAsteroid(cx+2, cy, childSize, a.vx+spread, a.vy)
	scene.Entities().Add(child1)
	scene.Entities().Add(child2)
}

func (a *Asteroid) IsDead() bool { return a.dead }

// --- Damageable ---

func (a *Asteroid) TakeDamage(amount int) {
	if a.dead {
		return
	}
	a.hp -= amount
	if a.hp <= 0 {
		a.hp = 0
		a.dead = true
	}
}

func (a *Asteroid) CurrentHP() int { return a.hp }
func (a *Asteroid) MaxHP() int     { return a.maxHP }

// --- Impulsable ---

func (a *Asteroid) ApplyImpulse(dvx, dvy float64) {
	mass := a.size.Mass()
	a.vx += dvx / mass
	a.vy += dvy / mass
	// Cap lateral speed to prevent infinite drift
	const maxVX = 8.0
	if a.vx > maxVX {
		a.vx = maxVX
	} else if a.vx < -maxVX {
		a.vx = -maxVX
	}
}

// CollidesWith implements precise collision for irregular asteroid shape
func (a *Asteroid) CollidesWith(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()

	// Get current rendered pixels
	pixels := a.sprite.Render()

	for row := range pixels {
		for col := range pixels[row] {
			if pixels[row][col].A > 0 { // Solid pixel
				px := a.x + col
				py := a.y + row

				// Check if this pixel overlaps with other entity's bounding box
				if px >= ox && px < ox+ow && py >= oy && py < oy+oh {
					return true
				}
			}
		}
	}
	return false
}

// craterInfo stores information about a crater for rendering
type craterInfo struct {
	cx, cy int
	radius int
}

// generateAsteroidSprite creates a procedural multi-colored asteroid
func generateAsteroidSprite(width, height int, size AsteroidSize) *draw.ColorMatrix {
	shape, craters := generateProceduralShape(width, height, size)

	matrix := make([][]draw.ColorKey, height)
	for i := range matrix {
		matrix[i] = make([]draw.ColorKey, width)
	}

	basePalette := generateRockPalette()

	colorCodes := draw.ColorMap{
		"0": {0, 0, 0, 0}, // Transparent
	}
	nextCode := 1 // Start from 1 for visible colors

	// Helper to convert int to ColorKey
	intToKey := func(n int) draw.ColorKey {
		// Use ASCII printable characters starting from space (32)
		// This gives us 95 different keys which should be more than enough
		return draw.ColorKey(string(rune(32 + n)))
	}

	for row := range shape {
		for col := range shape[row] {
			if !shape[row][col] {
				matrix[row][col] = "0" // Transparent
				continue
			}

			colorIndex := selectRockColor(row, col, width, height)

			for _, crater := range craters {
				dx := col - crater.cx
				dy := row - crater.cy
				distSq := dx*dx + dy*dy
				radiusSq := crater.radius * crater.radius

				if distSq < radiusSq {
					// Inside crater - shift to darker palette colors
					// Center of crater = darkest (0), edge = dark (1)
					intensity := float64(distSq) / float64(radiusSq) // 0 at center, 1 at edge

					if intensity < 0.5 {
						colorIndex = 0 // Darkest palette color in center
					} else {
						colorIndex = 1 // Dark palette color toward edges
					}
					break
				}
			}

			rockColor := basePalette[colorIndex]

			var existingCode draw.ColorKey
			for code, c := range colorCodes {
				if c == rockColor {
					existingCode = code
					break
				}
			}

			if existingCode != "" {
				matrix[row][col] = existingCode
			} else {
				key := intToKey(nextCode)
				colorCodes[key] = rockColor
				matrix[row][col] = key
				nextCode++
			}
		}
	}

	cm, err := draw.NewColorMatrix(matrix, &colorCodes, nil)
	if err != nil {
		return createFallbackAsteroid(width, height)
	}

	return cm
}

// generateProceduralShape creates an irregular asteroid shape and crater positions
func generateProceduralShape(width, height int, size AsteroidSize) ([][]bool, []craterInfo) {
	shape := make([][]bool, height)
	for i := range shape {
		shape[i] = make([]bool, width)
	}

	centerX := width / 2
	centerY := height / 2
	baseRadius := float64(width) / 2.5

	// Create irregular circular shape with noise
	for row := range shape {
		for col := range shape[row] {
			dx := float64(col - centerX)
			dy := float64(row - centerY)
			distance := (dx*dx + dy*dy)

			// Add noise to radius based on angle
			angle := float64(col+row*3) * 0.5
			radiusVariation := 0.7 + 0.3*noiseValue(angle)

			adjustedRadius := baseRadius * radiusVariation

			// Solid if within irregular radius
			shape[row][col] = distance < adjustedRadius*adjustedRadius
		}
	}

	// Generate craters for visual style (darker shaded areas)
	var craters []craterInfo
	if size >= AsteroidMedium {
		// More craters for larger asteroids
		numCraters := 1 + rand.Intn(3) // 1-3 craters
		if size >= AsteroidHuge {
			numCraters = 2 + rand.Intn(4) // 2-5 craters for huge+
		}

		for range numCraters {
			// Position craters away from edges for better visibility
			margin := width / 4
			cx := margin + rand.Intn(width-2*margin)
			cy := margin + rand.Intn(height-2*margin)

			// Scale crater size with asteroid size
			minRadius := 2
			maxRadius := 4
			if size >= AsteroidLarge {
				minRadius = 3
				maxRadius = 6
			}
			if size >= AsteroidHuge {
				minRadius = 4
				maxRadius = 8
			}

			craterRadius := minRadius + rand.Intn(maxRadius-minRadius+1)

			// Only add crater if it's within the asteroid shape
			if cx >= 0 && cx < width && cy >= 0 && cy < height && shape[cy][cx] {
				craters = append(craters, craterInfo{
					cx:     cx,
					cy:     cy,
					radius: craterRadius,
				})
			}
		}
	}

	return shape, craters
}

// generateRockPalette creates a palette of rock colors
func generateRockPalette() []color.RGBA {
	// Choose asteroid type: rocky (brown/orange) or metallic (gray/blue tint)
	asteroidType := rand.Float64()

	if asteroidType < 0.7 {
		// Rocky asteroid - warm brownish-orange tones (70% chance)
		baseOrange := 140 + rand.Intn(60) // 140-200

		return []color.RGBA{
			// Darkest (deep shadows)
			{
				R: uint8(baseOrange - 40),
				G: uint8(baseOrange - 60),
				B: uint8(baseOrange - 80),
				A: 255,
			},
			// Dark
			{
				R: uint8(baseOrange),
				G: uint8(baseOrange - 20),
				B: uint8(baseOrange - 40),
				A: 255,
			},
			// Medium (most common)
			{
				R: uint8(min(255, baseOrange+40)),
				G: uint8(baseOrange + 20),
				B: uint8(baseOrange - 10),
				A: 255,
			},
			// Light
			{
				R: uint8(min(255, baseOrange+70)),
				G: uint8(min(255, baseOrange+50)),
				B: uint8(baseOrange + 20),
				A: 255,
			},
			// Lightest (bright highlights) - edge glow
			{
				R: 255,
				G: uint8(min(255, baseOrange+80)),
				B: uint8(min(255, baseOrange+50)),
				A: 255,
			},
		}
	} else {
		// Metallic asteroid - cooler gray/blue tones (30% chance)
		baseGray := 120 + rand.Intn(60) // 120-180

		return []color.RGBA{
			// Darkest (shadows)
			{
				R: uint8(baseGray - 40),
				G: uint8(baseGray - 35),
				B: uint8(baseGray - 20),
				A: 255,
			},
			// Dark
			{
				R: uint8(baseGray),
				G: uint8(baseGray + 5),
				B: uint8(baseGray + 15),
				A: 255,
			},
			// Medium
			{
				R: uint8(baseGray + 30),
				G: uint8(baseGray + 35),
				B: uint8(min(255, baseGray+50)),
				A: 255,
			},
			// Light
			{
				R: uint8(min(255, baseGray+60)),
				G: uint8(min(255, baseGray+65)),
				B: uint8(min(255, baseGray+85)),
				A: 255,
			},
			// Lightest (metallic shine)
			{
				R: 240,
				G: 245,
				B: 255,
				A: 255,
			},
		}
	}
}

// selectRockColor chooses a color based on position (creates gradient effect)
func selectRockColor(row, col, width, height int) int {
	centerX := float64(width) / 2
	centerY := float64(height) / 2

	// Distance from center (for edge detection)
	dx := float64(col) - centerX
	dy := float64(row) - centerY
	distFromCenter := (dx*dx + dy*dy) / ((centerX * centerX) + (centerY * centerY))

	// Create lighting effect: top-left is lighter, bottom-right is darker
	lightScore := float64(width+height-row-col) / float64(width+height)

	// Add some randomness
	lightScore += (rand.Float64() - 0.5) * 0.3

	// Edge highlighting - outer edge gets bright color
	if distFromCenter > 0.85 {
		return 4 // Brightest on edges for visibility
	}

	// Clamp and map to palette index (0-4)
	if lightScore < 0.2 {
		return 0 // Darkest
	} else if lightScore < 0.4 {
		return 1 // Dark
	} else if lightScore < 0.6 {
		return 2 // Medium
	} else if lightScore < 0.8 {
		return 3 // Light
	} else {
		return 4 // Lightest
	}
}

// noiseValue provides simple pseudo-random variation
func noiseValue(x float64) float64 {
	// Simple deterministic noise function
	n := int(x * 12.9898)
	return float64((n*n*15731+789221)%1000) / 1000.0
}

// createFallbackAsteroid creates a simple single-color asteroid if generation fails
func createFallbackAsteroid(width, height int) *draw.ColorMatrix {
	matrix := make([][]draw.ColorKey, height)
	for i := range matrix {
		matrix[i] = make([]draw.ColorKey, width)
		for j := range matrix[i] {
			matrix[i][j] = "1" // All solid
		}
	}

	colors := draw.ColorMap{
		"1": {R: 100, G: 100, B: 100, A: 255},
	}

	cm, _ := draw.NewColorMatrix(matrix, &colors, nil)
	return cm
}
