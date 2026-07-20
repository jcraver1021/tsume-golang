package enemy

import (
	"math"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
	"tsumegolang/game/starshot/entity/effects"
)

const (
	mineFloatSpeed      = 1.0
	mineChasSpeed       = 2.5
	mineDetectionRadius = 160.0 // def.ScreenHeight / 4
	mineBlastRadius     = 72.0
	mineBlastDamage     = 999
	mineMaxHP           = 1
	mineValue           = 30
)

// Mine is a proximity-triggered explosive hazard. It drifts down the screen
// until the player strays within mineDetectionRadius, then chases and detonates.
type Mine struct {
	x, y          int
	fx, fy        float64
	width, height int

	idleSprite  *draw.ColorMatrix
	chaseSprite *draw.ColorMatrix

	chasing    bool
	dead       bool
	frameCount int
	maxFrames  int

	hp    int
	maxHP int
}

func NewMine(x, y int) (*Mine, error) {
	idle, err := loadMineSprite("mine_idle.yaml")
	if err != nil {
		return nil, err
	}

	chase, err := loadMineSprite("mine_chase.yaml")
	if err != nil {
		return nil, err
	}

	w, h := idle.Dimensions()

	return &Mine{
		x:           x - w/2,
		y:           y,
		fx:          float64(x - w/2),
		fy:          float64(y),
		width:       w,
		height:      h,
		idleSprite:  idle,
		chaseSprite: chase,
		hp:          mineMaxHP,
		maxHP:       mineMaxHP,
	}, nil
}

func loadMineSprite(filename string) (*draw.ColorMatrix, error) {
	data, err := spriteFiles.ReadFile("sprites/" + filename)
	if err != nil {
		return nil, err
	}
	return draw.ColorMatrixFromBytes(data)
}

func (m *Mine) Type() def.EntityType { return def.EntityTypeEnemy }

func (m *Mine) Location() (int, int) { return m.x, m.y }

func (m *Mine) Dimensions() (int, int) { return m.width, m.height }

func (m *Mine) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(m.x+m.width < ox || m.x > ox+ow || m.y+m.height < oy || m.y > oy+oh)
}

func (m *Mine) Act(scene def.Scene) {
	if m.dead {
		m.frameCount++
		return
	}

	players := scene.Entities().Get(def.EntityTypePlayer)
	if len(players) > 0 {
		px, py := players[0].Location()
		pw, ph := players[0].Dimensions()
		pcx := float64(px + pw/2)
		pcy := float64(py + ph/2)
		mcx := m.fx + float64(m.width)/2
		mcy := m.fy + float64(m.height)/2
		dx := pcx - mcx
		dy := pcy - mcy
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < mineDetectionRadius {
			m.chasing = true
		}

		if m.chasing && dist > 1 {
			m.fx += (dx / dist) * mineChasSpeed
			m.fy += (dy / dist) * mineChasSpeed
		}
	}

	if !m.chasing {
		m.fy += mineFloatSpeed
	}

	if m.fx < 0 {
		m.fx = 0
	}
	if m.fx+float64(m.width) > float64(scene.Width()) {
		m.fx = float64(scene.Width() - m.width)
	}

	m.x = int(m.fx)
	m.y = int(m.fy)
}

func (m *Mine) Draw(img *ebit.Image) {
	sprite := m.idleSprite
	if m.chasing {
		sprite = m.chaseSprite
	}
	pixels := sprite.Render()
	for row := range pixels {
		for col := range pixels[row] {
			c := pixels[row][col]
			if c.A > 0 {
				img.Set(m.x+col, m.y+row, c)
			}
		}
	}
}

func (m *Mine) CanBeRemoved() bool {
	if m.dead {
		return m.frameCount >= m.maxFrames
	}
	return m.y > def.ScreenHeight
}

// Mortal

func (m *Mine) GetDeathEffect() def.DeathEffect {
	return def.DeathEffect{
		SpawnVisualEffect: func(cx, cy int, scene def.Scene) {
			if exp, err := effects.NewExplosion(cx, cy, effects.ExplosionLarge); err == nil {
				scene.Entities().Add(exp)
			}
		},
	}
}

func (m *Mine) MarkAsDead(_ def.Scene) {
	m.dead = true
	m.frameCount = 0
	m.maxFrames = 60
}

func (m *Mine) IsDead() bool {
	return m.dead
}

// Damageable

func (m *Mine) TakeDamage(amount int) {
	if m.dead {
		return
	}
	m.hp -= amount
	if m.hp <= 0 {
		m.hp = 0
		m.dead = true
	}
}

func (m *Mine) CurrentHP() int {
	return m.hp
}

func (m *Mine) MaxHP() int {
	return m.maxHP
}

// Explosive

func (m *Mine) BlastRadius() float64 {
	return mineBlastRadius
}

func (m *Mine) BlastDamage() int {
	return mineBlastDamage
}

// Scorer

func (m *Mine) ScoreValue() int {
	return mineValue
}
