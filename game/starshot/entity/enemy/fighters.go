package enemy

// fighters.go — active enemy ships that pursue the player.
// Add new fighter types here; each section covers one ship and its brain.

import (
	"math"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
	"tsumegolang/game/starshot/entity/effects"
	"tsumegolang/game/starshot/entity/projectile"
)

const (
	enemyDrawScale = 2 // all enemies render at this scale via DrawImage
)

// ─── Chaser ───────────────────────────────────────────────────────────────────
// Chaser pursues the player and rams into them. It steers around asteroids but
// does not fire. See Hunter for a shooting variant.

const (
	chaserDefaultSpeed    = 2.0
	chaserDefaultTurnRate = 0.105 // ~6° per frame; full 180° in ~30 frames
	chaserLookahead       = 60    // px ahead to scan for obstacles (sensor range)
	chaserMaxHP           = 3
	chaserValue           = 10
)

type Chaser struct {
	x, y          int
	width, height int
	sprite        *draw.ColorMatrix
	cachedImg     *ebit.Image
	pixelBuf      []byte
	dead          bool
	frameCount    int
	maxFrames     int
	hp, maxHP     int
	fx, fy        float64
	brain         def.Brain
}

func NewChaser(x, y int) (*Chaser, error) {
	data, err := spriteFiles.ReadFile("sprites/chaser.yaml")
	if err != nil {
		return nil, err
	}
	sprite, err := draw.ColorMatrixFromBytes(data)
	if err != nil {
		return nil, err
	}
	w, h := sprite.Dimensions()
	return &Chaser{
		x:         x,
		y:         y,
		fx:        float64(x),
		fy:        float64(y),
		width:     int(float64(w) * enemyDrawScale),
		height:    int(float64(h) * enemyDrawScale),
		sprite:    sprite,
		cachedImg: ebit.NewImage(w, h),
		pixelBuf:  make([]byte, w*h*4),
		hp:        chaserMaxHP,
		maxHP:     chaserMaxHP,
		brain:     &ChaserBrain{Speed: chaserDefaultSpeed, TurnRate: chaserDefaultTurnRate},
	}, nil
}

func (c *Chaser) Type() def.EntityType {
	return def.EntityTypeEnemy
}

func (c *Chaser) Location() (int, int) {
	return c.x, c.y
}

func (c *Chaser) Dimensions() (int, int) {
	return c.width, c.height
}

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
	c.applyIntent(c.brain.Decide(c.Perceive(scene)), scene)
}

// Perceive emits SignalSelf, SignalPlayer, and SignalObstacle for hazards within
// chaserLookahead pixels along the current heading.
func (c *Chaser) Perceive(scene def.Scene) def.Perception {
	p := def.Perception{
		{Kind: def.SignalSelf, Condition: def.ConditionFor(c.hp, c.maxHP)},
	}
	players := scene.Entities().Get(def.EntityTypePlayer)
	if len(players) == 0 {
		return p
	}
	px, py := players[0].Location()
	pw, ph := players[0].Dimensions()
	myCX := c.fx + float64(c.width)/2
	myCY := c.fy + float64(c.height)/2
	dx := float64(px+pw/2) - myCX
	dy := float64(py+ph/2) - myCY
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return p
	}
	nx, ny := dx/dist, dy/dist
	playerCondition := def.ConditionHealthy
	if d, ok := players[0].(def.Damageable); ok {
		playerCondition = def.ConditionFor(d.CurrentHP(), d.MaxHP())
	}
	p = append(p, def.Signal{
		Kind:      def.SignalPlayer,
		Direction: [2]float64{nx, ny},
		Distance:  dist,
		Condition: playerCondition,
	})
	for _, obs := range scene.Entities().Get(def.EntityTypeObstacle) {
		ox, oy := obs.Location()
		ow, oh := obs.Dimensions()
		obsCX := float64(ox+ow/2) - myCX
		obsCY := float64(oy+oh/2) - myCY
		along := obsCX*nx + obsCY*ny
		if along < 0 || along > float64(chaserLookahead) {
			continue
		}
		perpX := obsCX - along*nx
		perpY := obsCY - along*ny
		if math.Sqrt(perpX*perpX+perpY*perpY) >= float64(ow/2+c.width/2)*1.4 {
			continue
		}
		obsDist := math.Sqrt(obsCX*obsCX + obsCY*obsCY)
		var obsDir [2]float64
		if obsDist > 0 {
			obsDir = [2]float64{obsCX / obsDist, obsCY / obsDist}
		}
		p = append(p, def.Signal{Kind: def.SignalObstacle, Direction: obsDir, Distance: obsDist})
	}
	return p
}

func (c *Chaser) applyIntent(intent def.Intent, scene def.Scene) {
	c.fx += intent.Direction[0] * intent.Speed
	c.fy += intent.Direction[1] * intent.Speed
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
	draw.DrawScaled(img, c.cachedImg, c.pixelBuf, c.sprite, float64(c.x), float64(c.y), enemyDrawScale)
}

func (c *Chaser) CanBeRemoved() bool {
	return c.dead && c.frameCount >= c.maxFrames
}

func (c *Chaser) GetDeathEffect() def.DeathEffect {
	return def.DeathEffect{
		SpawnVisualEffect: func(cx, cy int, scene def.Scene) {
			if exp, err := effects.NewExplosion(cx, cy, effects.ExplosionMedium); err == nil {
				scene.Entities().Add(exp)
			}
		},
	}
}

func (c *Chaser) MarkAsDead(_ def.Scene) {
	c.dead = true
	c.frameCount = 0
	c.maxFrames = 60
}

func (c *Chaser) IsDead() bool {
	return c.dead
}

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

// ─── ChaserBrain ──────────────────────────────────────────────────────────────
// ChaserBrain steers toward the player at constant speed, turning to avoid
// obstacles. Three steering choices per frame: left, right, or aim-at-player.

type ChaserBrain struct {
	Speed    float64
	TurnRate float64
	heading  [2]float64
}

func (b *ChaserBrain) Decide(p def.Perception) def.Intent {
	if b.heading[0] == 0 && b.heading[1] == 0 {
		b.heading = [2]float64{0, 1}
	}
	var playerSignal *def.Signal
	var threatSignal float64
	for i := range p {
		s := &p[i]
		switch s.Kind {
		case def.SignalPlayer:
			playerSignal = s
		case def.SignalObstacle:
			cross := b.heading[0]*s.Direction[1] - b.heading[1]*s.Direction[0]
			if cross > 0 {
				threatSignal -= 1
			} else {
				threatSignal += 1
			}
		}
	}
	if playerSignal == nil {
		return def.Intent{}
	}
	var target [2]float64
	switch {
	case threatSignal > 0:
		target = rotate2D(b.heading, b.TurnRate)
	case threatSignal < 0:
		target = rotate2D(b.heading, -b.TurnRate)
	default:
		target = playerSignal.Direction
	}
	b.heading = rotateToward(b.heading, target, b.TurnRate)
	return def.Intent{Direction: b.heading, Speed: b.Speed}
}

// ─── Hunter ───────────────────────────────────────────────────────────────────
// Hunter pursues the player and fires when aligned. Steers around asteroids.
// See Chaser for a non-firing ramming variant.

const (
	hunterDefaultSpeed    = 1.8
	hunterDefaultTurnRate = 0.08
	hunterLookahead       = 60
	hunterMaxHP           = 4
	hunterValue           = 25
	hunterFireRate        = 75 // frames between shots (~1.25s at 60fps)
)

type Hunter struct {
	x, y          int
	width, height int
	sprite        *draw.ColorMatrix
	cachedImg     *ebit.Image
	pixelBuf      []byte
	dead          bool
	frameCount    int
	maxFrames     int
	hp, maxHP     int
	fx, fy        float64
	fireCooldown  int
	brain         def.Brain
}

func NewHunter(x, y int) (*Hunter, error) {
	data, err := spriteFiles.ReadFile("sprites/hunter.yaml")
	if err != nil {
		return nil, err
	}
	sprite, err := draw.ColorMatrixFromBytes(data)
	if err != nil {
		return nil, err
	}
	w, h := sprite.Dimensions()
	return &Hunter{
		x:         x,
		y:         y,
		fx:        float64(x),
		fy:        float64(y),
		width:     int(float64(w) * enemyDrawScale),
		height:    int(float64(h) * enemyDrawScale),
		sprite:    sprite,
		cachedImg: ebit.NewImage(w, h),
		pixelBuf:  make([]byte, w*h*4),
		hp:        hunterMaxHP,
		maxHP:     hunterMaxHP,
		brain:     &HunterBrain{Speed: hunterDefaultSpeed, TurnRate: hunterDefaultTurnRate},
	}, nil
}

func (h *Hunter) Type() def.EntityType {
	return def.EntityTypeEnemy
}

func (h *Hunter) Location() (int, int) {
	return h.x, h.y
}

func (h *Hunter) Dimensions() (int, int) {
	return h.width, h.height
}

func (h *Hunter) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(h.x+h.width < ox || h.x > ox+ow || h.y+h.height < oy || h.y > oy+oh)
}

func (h *Hunter) Act(scene def.Scene) {
	if h.dead {
		h.frameCount++
		return
	}
	h.applyIntent(h.brain.Decide(h.Perceive(scene)), scene)
}

// Perceive emits SignalSelf, SignalPlayer, and SignalObstacle for hazards within
// hunterLookahead pixels along the current heading.
func (h *Hunter) Perceive(scene def.Scene) def.Perception {
	p := def.Perception{
		{Kind: def.SignalSelf, Condition: def.ConditionFor(h.hp, h.maxHP)},
	}
	players := scene.Entities().Get(def.EntityTypePlayer)
	if len(players) == 0 {
		return p
	}
	px, py := players[0].Location()
	pw, ph := players[0].Dimensions()
	myCX := h.fx + float64(h.width)/2
	myCY := h.fy + float64(h.height)/2
	dx := float64(px+pw/2) - myCX
	dy := float64(py+ph/2) - myCY
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 1 {
		return p
	}
	nx, ny := dx/dist, dy/dist
	playerCondition := def.ConditionHealthy
	if d, ok := players[0].(def.Damageable); ok {
		playerCondition = def.ConditionFor(d.CurrentHP(), d.MaxHP())
	}
	p = append(p, def.Signal{
		Kind:      def.SignalPlayer,
		Direction: [2]float64{nx, ny},
		Distance:  dist,
		Condition: playerCondition,
	})
	for _, obs := range scene.Entities().Get(def.EntityTypeObstacle) {
		ox, oy := obs.Location()
		ow, oh := obs.Dimensions()
		obsCX := float64(ox+ow/2) - myCX
		obsCY := float64(oy+oh/2) - myCY
		along := obsCX*nx + obsCY*ny
		if along < 0 || along > float64(hunterLookahead) {
			continue
		}
		perpX := obsCX - along*nx
		perpY := obsCY - along*ny
		if math.Sqrt(perpX*perpX+perpY*perpY) >= float64(ow/2+h.width/2)*1.4 {
			continue
		}
		obsDist := math.Sqrt(obsCX*obsCX + obsCY*obsCY)
		var obsDir [2]float64
		if obsDist > 0 {
			obsDir = [2]float64{obsCX / obsDist, obsCY / obsDist}
		}
		p = append(p, def.Signal{Kind: def.SignalObstacle, Direction: obsDir, Distance: obsDist})
	}
	return p
}

func (h *Hunter) applyIntent(intent def.Intent, scene def.Scene) {
	if h.fireCooldown > 0 {
		h.fireCooldown--
	}
	if intent.Fire && h.fireCooldown == 0 {
		cx := h.x + h.width/2
		cy := h.y + h.height/2
		scene.Entities().Add(projectile.NewEnemyBullet(cx, cy, intent.FireAim))
		h.fireCooldown = hunterFireRate
	}
	h.fx += intent.Direction[0] * intent.Speed
	h.fy += intent.Direction[1] * intent.Speed
	if h.fx < 0 {
		h.fx = 0
	}
	if h.fy < 0 {
		h.fy = 0
	}
	if h.fx+float64(h.width) > float64(scene.Width()) {
		h.fx = float64(scene.Width() - h.width)
	}
	if h.fy+float64(h.height) > float64(scene.Height()) {
		h.fy = float64(scene.Height() - h.height)
	}
	h.x = int(h.fx)
	h.y = int(h.fy)
}

func (h *Hunter) Draw(img *ebit.Image) {
	draw.DrawScaled(img, h.cachedImg, h.pixelBuf, h.sprite, float64(h.x), float64(h.y), enemyDrawScale)
}

func (h *Hunter) CanBeRemoved() bool {
	return h.dead && h.frameCount >= h.maxFrames
}

func (h *Hunter) GetDeathEffect() def.DeathEffect {
	return def.DeathEffect{
		SpawnVisualEffect: func(cx, cy int, scene def.Scene) {
			if exp, err := effects.NewExplosion(cx, cy, effects.ExplosionMedium); err == nil {
				scene.Entities().Add(exp)
			}
		},
	}
}

func (h *Hunter) MarkAsDead(_ def.Scene) {
	h.dead = true
	h.frameCount = 0
	h.maxFrames = 60
}

func (h *Hunter) IsDead() bool {
	return h.dead
}

func (h *Hunter) TakeDamage(amount int) {
	if h.dead {
		return
	}
	h.hp -= amount
	if h.hp <= 0 {
		h.hp = 0
		h.dead = true
	}
}

func (h *Hunter) CurrentHP() int {
	return h.hp
}

func (h *Hunter) MaxHP() int {
	return h.maxHP
}

func (h *Hunter) ScoreValue() int {
	return hunterValue
}

// ─── HunterBrain ──────────────────────────────────────────────────────────────
// HunterBrain steers like ChaserBrain and fires when the heading is closely
// aligned with the player (within ~18°, dot > 0.95).

type HunterBrain struct {
	Speed    float64
	TurnRate float64
	heading  [2]float64
}

func (b *HunterBrain) Decide(p def.Perception) def.Intent {
	if b.heading[0] == 0 && b.heading[1] == 0 {
		b.heading = [2]float64{0, 1}
	}
	var playerSignal *def.Signal
	var threatSignal float64
	for i := range p {
		s := &p[i]
		switch s.Kind {
		case def.SignalPlayer:
			playerSignal = s
		case def.SignalObstacle:
			cross := b.heading[0]*s.Direction[1] - b.heading[1]*s.Direction[0]
			if cross > 0 {
				threatSignal -= 1
			} else {
				threatSignal += 1
			}
		}
	}
	if playerSignal == nil {
		return def.Intent{}
	}
	var target [2]float64
	switch {
	case threatSignal > 0:
		target = rotate2D(b.heading, b.TurnRate)
	case threatSignal < 0:
		target = rotate2D(b.heading, -b.TurnRate)
	default:
		target = playerSignal.Direction
	}
	b.heading = rotateToward(b.heading, target, b.TurnRate)
	intent := def.Intent{Direction: b.heading, Speed: b.Speed}
	dot := b.heading[0]*playerSignal.Direction[0] + b.heading[1]*playerSignal.Direction[1]
	if dot > 0.95 {
		intent.Fire = true
		intent.FireAim = playerSignal.Direction
	}
	return intent
}

// ─── Steering utilities ───────────────────────────────────────────────────────

// rotate2D rotates a 2D vector by angle radians (positive = clockwise in screen space).
func rotate2D(v [2]float64, angle float64) [2]float64 {
	cos, sin := math.Cos(angle), math.Sin(angle)
	return [2]float64{v[0]*cos - v[1]*sin, v[0]*sin + v[1]*cos}
}

// rotateToward rotates current toward target by at most maxAngle radians,
// choosing the shorter arc via the cross product.
func rotateToward(current, target [2]float64, maxAngle float64) [2]float64 {
	dot := math.Max(-1, math.Min(1, current[0]*target[0]+current[1]*target[1]))
	if dot >= 1 {
		return target
	}
	angle := math.Acos(dot)
	if angle <= maxAngle {
		return target
	}
	cross := current[0]*target[1] - current[1]*target[0]
	sign := 1.0
	if cross < 0 {
		sign = -1.0
	}
	return rotate2D(current, sign*maxAngle)
}
