package enemy

import (
	"math"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
)

const (
	chaserSpeed      = 2
	chaserLookahead  = 60 // px ahead to scan for obstacles
	chaserDodgeForce = 3  // lateral push when dodging
	chaserMaxHP      = 3
	chaserValue      = 10
)

// Chaser is a hostile ship that pursues the player and steers around asteroids.
type Chaser struct {
	x, y          int
	width, height int
	sprite        *draw.ColorMatrix
	dead          bool
	frameCount    int
	maxFrames     int

	hp    int
	maxHP int

	// fractional sub-pixel position for smooth diagonal movement
	fx, fy float64
}

func NewChaser(x, y int) (*Chaser, error) {
	sprite := generateChaserSprite()
	w, h := sprite.Dimensions()
	return &Chaser{
		x:      x,
		y:      y,
		fx:     float64(x),
		fy:     float64(y),
		width:  w,
		height: h,
		sprite: sprite,
		hp:     chaserMaxHP,
		maxHP:  chaserMaxHP,
	}, nil
}

func (c *Chaser) Type() def.EntityType { return def.EntityTypeEnemy }

func (c *Chaser) Location() (int, int) { return c.x, c.y }

func (c *Chaser) Dimensions() (int, int) { return c.width, c.height }

func (c *Chaser) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(c.x+c.width < ox || c.x > ox+ow || c.y+c.height < oy || c.y > oy+oh)
}

func (c *Chaser) Act(scene def.Scene) {
	if c.dead {
		c.frameCount++
		return
	}

	// Find the player to chase
	players := scene.Entities().Get(def.EntityTypePlayer)
	if len(players) == 0 {
		return
	}
	px, py := players[0].Location()
	pw, ph := players[0].Dimensions()
	targetX := float64(px + pw/2)
	targetY := float64(py + ph/2)

	myCX := c.fx + float64(c.width)/2
	myCY := c.fy + float64(c.height)/2

	dx := targetX - myCX
	dy := targetY - myCY
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return
	}
	// Normalized direction toward player
	nx := dx / dist
	ny := dy / dist

	// Dodge: scan obstacles within lookahead ahead of the chaser
	obstacles := scene.Entities().Get(def.EntityTypeObstacle)
	lateralPush := 0.0
	for _, obs := range obstacles {
		ox, oy := obs.Location()
		ow, oh := obs.Dimensions()
		obsCX := float64(ox + ow/2)
		obsCY := float64(oy + oh/2)

		// Project obstacle center onto chaser's movement direction
		toObsX := obsCX - myCX
		toObsY := obsCY - myCY
		along := toObsX*nx + toObsY*ny
		if along < 0 || along > float64(chaserLookahead) {
			continue
		}
		// Perpendicular distance from movement line
		perpX := toObsX - along*nx
		perpY := toObsY - along*ny
		perpDist := math.Sqrt(perpX*perpX + perpY*perpY)

		dangerRadius := float64(ow/2+c.width/2) * 1.4
		if perpDist < dangerRadius {
			// Push laterally away from obstacle
			// The perpendicular direction (left normal of nx,ny) is (-ny, nx)
			// Determine sign: if obstacle is to the right, push left
			cross := toObsX*(-ny) - toObsY*(-nx) // sign of cross product
			if cross < 0 {
				lateralPush += chaserDodgeForce
			} else {
				lateralPush -= chaserDodgeForce
			}
		}
	}

	// Left normal of movement direction: (-ny, nx)
	moveX := nx*chaserSpeed + (-ny)*lateralPush
	moveY := ny*chaserSpeed + nx*lateralPush

	c.fx += moveX
	c.fy += moveY

	// Clamp to screen
	if c.fx < 0 {
		c.fx = 0
	}
	if c.fy < 0 {
		c.fy = 0
	}
	if c.fx+float64(c.width) > float64(scene.Width()) {
		c.fx = float64(scene.Width() - c.width)
	}
	if c.fy+float64(c.height) > float64(scene.Height()) {
		c.fy = float64(scene.Height() - c.height)
	}

	c.x = int(c.fx)
	c.y = int(c.fy)
}

func (c *Chaser) Draw(img *ebit.Image) {
	pixels := c.sprite.Render()
	for row := range pixels {
		for col := range pixels[row] {
			color := pixels[row][col]
			if color.A > 0 {
				img.Set(c.x+col, c.y+row, color)
			}
		}
	}
}

func (c *Chaser) CanBeRemoved() bool {
	return c.dead && c.frameCount >= c.maxFrames
}

// Mortal interface

func (c *Chaser) GetDeathEffect() def.DeathEffect {
	return def.DeathEffect{
		ExplosionSize:      def.ExplosionMedium,
		SlowdownMultiplier: 0,
		SlowdownDuration:   0,
	}
}

func (c *Chaser) MarkAsDead(scene def.Scene) {
	c.dead = true
	c.frameCount = 0
	c.maxFrames = 60 // time for explosion to play
}

func (c *Chaser) IsDead() bool { return c.dead }

// --- Damageable ---

func (c *Chaser) TakeDamage(amount int) {
	if c.dead {
		return
	}
	c.hp -= amount
	if c.hp <= 0 {
		c.hp = 0
		c.dead = true
	}
}

func (c *Chaser) CurrentHP() int {
	return c.hp
}

func (c *Chaser) MaxHP() int {
	return c.maxHP
}

func (c *Chaser) ScoreValue() int {
	return chaserValue
}

// generateChaserSprite builds a 14×14 hostile arrowhead ship pointing downward.
func generateChaserSprite() *draw.ColorMatrix {
	// Design (14 wide, 16 tall): downward-pointing angular craft
	// X = hull body, E = engine glow, C = cockpit, 0 = transparent
	rows := []string{
		"00000XX00000XX", // 0  wing tips
		"00000XXXX00XX0", // 1
		"00000XXXXXX000", // 2  upper wings
		"0000XXXXXXX000", // 3
		"000XXXXXXXXX00", // 4  mid hull
		"00XXXXXXXXXXX0", // 5
		"0XXXXXXXXXXXXX", // 6  full width
		"0XCCCXXXXXXXXX", // 7  cockpit
		"0XCCCXXXXXXXXX", // 8
		"0XXXXXXXXXXXXX", // 9
		"00XXXXXXXXXXX0", // 10
		"000XXXXXXXXX00", // 11 narrowing
		"0000XXXXXXX000", // 12
		"00000EEEEE0000", // 13 engine glow
		"000000EEE00000", // 14
		"0000000E000000", // 15 engine core
	}

	colors := draw.ColorMap{
		"0": {0, 0, 0, 0},
		"X": {180, 30, 30, 255},  // dark red hull
		"C": {255, 100, 50, 255}, // orange cockpit
		"E": {255, 200, 50, 255}, // yellow engine glow
	}

	matrix := make([][]draw.ColorKey, len(rows))
	for r, row := range rows {
		matrix[r] = make([]draw.ColorKey, len(row))
		for col, ch := range row {
			matrix[r][col] = draw.ColorKey(string(ch))
		}
	}

	cm, err := draw.NewColorMatrix(matrix, &colors, nil)
	if err != nil {
		// Fallback: solid red rectangle
		fb := make([][]draw.ColorKey, 16)
		fbc := draw.ColorMap{"X": {180, 30, 30, 255}}
		for r := range fb {
			fb[r] = make([]draw.ColorKey, 14)
			for col := range fb[r] {
				fb[r][col] = "X"
			}
		}
		cm, _ = draw.NewColorMatrix(fb, &fbc, nil)
	}
	return cm
}
