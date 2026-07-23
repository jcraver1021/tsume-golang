package enemy

// mines.go — explosive hazards that drift or follow paths and detonate on contact
// or proximity. Add new mine types here; each section covers one type and its brain.

import (
	"math"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
	"tsumegolang/game/starshot/entity/effects"

	ebit "github.com/hajimehoshi/ebiten/v2"
)

// loadMineSprite reads a mine sprite YAML from the embedded sprites directory.
func loadMineSprite(filename string) (*draw.ColorMatrix, error) {
	data, err := spriteFiles.ReadFile("sprites/" + filename)
	if err != nil {
		return nil, err
	}
	return draw.ColorMatrixFromBytes(data)
}

// PathSegment defines an additional velocity applied for a given number of frames.
// VX and VY are added on top of the mine's downward drift while this segment is active.
type PathSegment struct {
	Frames int
	VX, VY float64
}

// ─── MineBrain ────────────────────────────────────────────────────────────────
// MineBrain drifts until a player signal is perceived, then chases directly.
// Detection and tracking are handled by Mine.Perceive; this brain only decides speed.

type MineBrain struct {
	ChaseSpeed float64
	DriftSpeed float64
}

func (b *MineBrain) Decide(p def.Perception) def.Intent {
	for _, s := range p {
		if s.Kind == def.SignalPlayer && s.Distance > 1 {
			return def.Intent{Direction: s.Direction, Speed: b.ChaseSpeed}
		}
	}
	return def.Intent{Direction: [2]float64{0, 1}, Speed: b.DriftSpeed}
}

// ─── Mine (contact mine) ──────────────────────────────────────────────────────
// Mine drifts slowly downward until the player enters its detection radius,
// then chases and detonates on direct contact with the player.

const (
	mineDefaultDriftSpeed = 1.0
	mineDefaultChaseSpeed = 2.5
	mineDetectionRadius   = 160.0 // px to player before locking on
	mineBlastRadius       = 72.0
	mineBlastDamage       = 999
	mineMaxHP             = 1
	mineValue             = 30
)

type Mine struct {
	x, y          int
	fx, fy        float64
	drift         float64
	width, height int
	idleSprite    *draw.ColorMatrix
	chaseSprite   *draw.ColorMatrix
	cachedImg     *ebit.Image
	pixelBuf      []byte
	chasing       bool
	dead          bool
	frameCount    int
	maxFrames     int
	hp, maxHP     int
	brain         def.Brain
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
	scaledW := int(float64(w) * enemyDrawScale)
	scaledH := int(float64(h) * enemyDrawScale)
	startX := x - scaledW/2
	return &Mine{
		x:           startX,
		y:           y,
		fx:          float64(startX),
		fy:          float64(y),
		drift:       mineDefaultDriftSpeed,
		width:       scaledW,
		height:      scaledH,
		idleSprite:  idle,
		chaseSprite: chase,
		cachedImg:   ebit.NewImage(w, h),
		pixelBuf:    make([]byte, w*h*4),
		hp:          mineMaxHP,
		maxHP:       mineMaxHP,
		brain:       &MineBrain{ChaseSpeed: mineDefaultChaseSpeed, DriftSpeed: mineDefaultDriftSpeed},
	}, nil
}

func (m *Mine) Type() def.EntityType {
	return def.EntityTypeEnemy
}

func (m *Mine) Location() (int, int) {
	return m.x, m.y
}

func (m *Mine) Dimensions() (int, int) {
	return m.width, m.height
}

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
	perception := m.Perceive(scene)
	for _, s := range perception {
		if s.Kind == def.SignalPlayer {
			m.chasing = true
			break
		}
	}
	m.applyIntent(m.brain.Decide(perception), scene)
}

// Perceive emits SignalSelf and, if the player is within mineDetectionRadius or
// already locked on, SignalPlayer.
func (m *Mine) Perceive(scene def.Scene) def.Perception {
	p := def.Perception{
		{Kind: def.SignalSelf, Condition: def.ConditionFor(m.hp, m.maxHP)},
	}
	players := scene.Entities().Get(def.EntityTypePlayer)
	if len(players) == 0 {
		return p
	}
	px, py := players[0].Location()
	pw, ph := players[0].Dimensions()
	mcx := m.fx + float64(m.width)/2
	mcy := m.fy + float64(m.height)/2
	dx := float64(px+pw/2) - mcx
	dy := float64(py+ph/2) - mcy
	dist := math.Sqrt(dx*dx + dy*dy)
	if (dist < mineDetectionRadius || m.chasing) && dist >= 1 {
		playerCondition := def.ConditionHealthy
		if d, ok := players[0].(def.Damageable); ok {
			playerCondition = def.ConditionFor(d.CurrentHP(), d.MaxHP())
		}
		p = append(p, def.Signal{
			Kind:      def.SignalPlayer,
			Direction: [2]float64{dx / dist, dy / dist},
			Distance:  dist,
			Condition: playerCondition,
		})
	}
	return p
}

func (m *Mine) applyIntent(intent def.Intent, scene def.Scene) {
	m.fx += intent.Direction[0] * intent.Speed
	m.fy += intent.Direction[1] * intent.Speed
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
	draw.DrawScaled(img, m.cachedImg, m.pixelBuf, sprite, float64(m.x), float64(m.y), enemyDrawScale)
}

func (m *Mine) SetDrift(drift float64) {
	m.drift = drift
	if brain, ok := m.brain.(*MineBrain); ok {
		brain.DriftSpeed = drift
	}
}

func (m *Mine) GetDrift() float64 {
	return m.drift
}

func (m *Mine) ResetDrift() {
	m.SetDrift(mineDefaultDriftSpeed)
}

func (m *Mine) CanBeRemoved() bool {
	if m.dead {
		return m.frameCount >= m.maxFrames
	}
	return m.y > def.ScreenHeight
}

func (m *Mine) GetDeathEffect() def.DeathEffect {
	return def.DeathEffect{
		SpawnVisualEffect: func(cx, cy int, scene def.Scene) {
			if exp, err := effects.NewExplosionScaled(cx, cy, effects.ExplosionLarge, 3.0); err == nil {
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

func (m *Mine) BlastRadius() float64 {
	return mineBlastRadius
}

func (m *Mine) BlastDamage() int {
	return mineBlastDamage
}

func (m *Mine) ScoreValue() int {
	return mineValue
}

// ─── RangeMine (proximity countdown) ─────────────────────────────────────────
// RangeMine drifts slowly downward. When the player lingers within
// rangeMineDetectionRadius for rangeMineDetonateFrames consecutive frames, it
// detonates. Its orange lights flash rapidly once active. Also detonates on contact.
// The detection radius is wider than the blast radius so a player who triggers the
// countdown can survive by staying between the two radii.

const (
	rangeMineDetectionRadius = 120.0 // wider than blast so player can escape
	rangeMineDetonateFrames  = 10
	rangeMineBlastRadius     = 72.0
	rangeMineBlastDamage     = 999
	rangeMineMaxHP           = 1
	rangeMineValue           = 50
	rangeMineDrift           = 0.5
)

type RangeMine struct {
	x, y            int
	fx, fy          float64
	drift           float64
	width, height   int
	idleSprite      *draw.ColorMatrix
	activeSprite    *draw.ColorMatrix
	cachedImg       *ebit.Image
	pixelBuf        []byte
	proximityFrames int
	active          bool
	dead            bool
	frameCount      int
	maxFrames       int
	hp, maxHP       int
}

func NewRangeMine(x, y int) (*RangeMine, error) {
	idle, err := loadMineSprite("range_mine_idle.yaml")
	if err != nil {
		return nil, err
	}
	active, err := loadMineSprite("range_mine_active.yaml")
	if err != nil {
		return nil, err
	}
	w, h := idle.Dimensions()
	scaledW := int(float64(w) * enemyDrawScale)
	scaledH := int(float64(h) * enemyDrawScale)
	startX := x - scaledW/2
	return &RangeMine{
		x:            startX,
		y:            y,
		fx:           float64(startX),
		fy:           float64(y),
		drift:        rangeMineDrift,
		width:        scaledW,
		height:       scaledH,
		idleSprite:   idle,
		activeSprite: active,
		cachedImg:    ebit.NewImage(w, h),
		pixelBuf:     make([]byte, w*h*4),
		hp:           rangeMineMaxHP,
		maxHP:        rangeMineMaxHP,
	}, nil
}

func (r *RangeMine) Type() def.EntityType {
	return def.EntityTypeEnemy
}

func (r *RangeMine) Location() (int, int) {
	return r.x, r.y
}

func (r *RangeMine) Dimensions() (int, int) {
	return r.width, r.height
}

func (r *RangeMine) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(r.x+r.width < ox || r.x > ox+ow || r.y+r.height < oy || r.y > oy+oh)
}

func (r *RangeMine) Act(scene def.Scene) {
	if r.dead {
		r.frameCount++
		return
	}
	r.fy += r.drift
	r.x = int(r.fx)
	r.y = int(r.fy)
	players := scene.Entities().Get(def.EntityTypePlayer)
	if len(players) == 0 {
		r.proximityFrames = 0
		r.active = false
		return
	}
	px, py := players[0].Location()
	pw, ph := players[0].Dimensions()
	mcx := r.fx + float64(r.width)/2
	mcy := r.fy + float64(r.height)/2
	dx := float64(px+pw/2) - mcx
	dy := float64(py+ph/2) - mcy
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist <= rangeMineDetectionRadius {
		r.proximityFrames++
		r.active = true
	} else {
		r.proximityFrames = 0
		r.active = false
	}
}

func (r *RangeMine) ReadyToDetonate() bool {
	return r.proximityFrames >= rangeMineDetonateFrames
}

func (r *RangeMine) SetDrift(drift float64) {
	r.drift = drift
}

func (r *RangeMine) GetDrift() float64 {
	return r.drift
}

func (r *RangeMine) ResetDrift() {
	r.drift = rangeMineDrift
}

func (r *RangeMine) Draw(img *ebit.Image) {
	sprite := r.idleSprite
	if r.active {
		sprite = r.activeSprite
	}
	draw.DrawScaled(img, r.cachedImg, r.pixelBuf, sprite, float64(r.x), float64(r.y), enemyDrawScale)
}

func (r *RangeMine) CanBeRemoved() bool {
	if r.dead {
		return r.frameCount >= r.maxFrames
	}
	return r.y > def.ScreenHeight
}

func (r *RangeMine) GetDeathEffect() def.DeathEffect {
	return def.DeathEffect{
		SpawnVisualEffect: func(cx, cy int, scene def.Scene) {
			if exp, err := effects.NewExplosionScaled(cx, cy, effects.ExplosionLarge, 3.0); err == nil {
				scene.Entities().Add(exp)
			}
		},
	}
}

func (r *RangeMine) MarkAsDead(_ def.Scene) {
	r.dead = true
	r.frameCount = 0
	r.maxFrames = 60
}

func (r *RangeMine) IsDead() bool {
	return r.dead
}

func (r *RangeMine) TakeDamage(amount int) {
	if r.dead {
		return
	}
	r.hp -= amount
	if r.hp <= 0 {
		r.hp = 0
		r.dead = true
	}
}

func (r *RangeMine) CurrentHP() int {
	return r.hp
}

func (r *RangeMine) MaxHP() int {
	return r.maxHP
}

func (r *RangeMine) BlastRadius() float64 {
	return rangeMineBlastRadius
}

func (r *RangeMine) BlastDamage() int {
	return rangeMineBlastDamage
}

func (r *RangeMine) ScoreValue() int {
	return rangeMineValue
}

// ─── PathMine (path-following, contact detonation) ────────────────────────────
// PathMine follows a repeating sequence of PathSegments while drifting downward.
// It has a slow blue pulsing light and detonates only on direct contact.

const (
	pathMineDrift       = 0.8
	pathMineBlastRadius = 72.0
	pathMineBlastDamage = 999
	pathMineMaxHP       = 1
	pathMineValue       = 20
)

type PathMine struct {
	x, y          int
	fx, fy        float64
	drift         float64
	width, height int
	sprite        *draw.ColorMatrix
	cachedImg     *ebit.Image
	pixelBuf      []byte
	path          []PathSegment
	pathFrame     int
	segmentTick   int
	dead          bool
	frameCount    int
	maxFrames     int
	hp, maxHP     int
}

func NewPathMine(x, y int, path []PathSegment) (*PathMine, error) {
	data, err := spriteFiles.ReadFile("sprites/path_mine.yaml")
	if err != nil {
		return nil, err
	}
	sprite, err := draw.ColorMatrixFromBytes(data)
	if err != nil {
		return nil, err
	}
	w, h := sprite.Dimensions()
	scaledW := int(float64(w) * enemyDrawScale)
	scaledH := int(float64(h) * enemyDrawScale)
	startX := x - scaledW/2
	return &PathMine{
		x:         startX,
		y:         y,
		fx:        float64(startX),
		fy:        float64(y),
		drift:     pathMineDrift,
		width:     scaledW,
		height:    scaledH,
		sprite:    sprite,
		cachedImg: ebit.NewImage(w, h),
		pixelBuf:  make([]byte, w*h*4),
		path:      path,
		hp:        pathMineMaxHP,
		maxHP:     pathMineMaxHP,
	}, nil
}

func (p *PathMine) Type() def.EntityType {
	return def.EntityTypeEnemy
}

func (p *PathMine) Location() (int, int) {
	return p.x, p.y
}

func (p *PathMine) Dimensions() (int, int) {
	return p.width, p.height
}

func (p *PathMine) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(p.x+p.width < ox || p.x > ox+ow || p.y+p.height < oy || p.y > oy+oh)
}

func (p *PathMine) Act(scene def.Scene) {
	if p.dead {
		p.frameCount++
		return
	}
	p.fy += p.drift
	if len(p.path) > 0 {
		seg := p.path[p.pathFrame]
		p.fx += seg.VX
		p.fy += seg.VY
		p.segmentTick++
		if p.segmentTick >= seg.Frames {
			p.segmentTick = 0
			p.pathFrame = (p.pathFrame + 1) % len(p.path)
		}
	}
	if p.fx < 0 {
		p.fx = 0
	}
	if p.fx+float64(p.width) > float64(scene.Width()) {
		p.fx = float64(scene.Width() - p.width)
	}
	p.x = int(p.fx)
	p.y = int(p.fy)
}

func (p *PathMine) Draw(img *ebit.Image) {
	draw.DrawScaled(img, p.cachedImg, p.pixelBuf, p.sprite, float64(p.x), float64(p.y), enemyDrawScale)
}

func (p *PathMine) SetDrift(drift float64) {
	p.drift = drift
}

func (p *PathMine) GetDrift() float64 {
	return p.drift
}

func (p *PathMine) ResetDrift() {
	p.drift = pathMineDrift
}

func (p *PathMine) CanBeRemoved() bool {
	if p.dead {
		return p.frameCount >= p.maxFrames
	}
	return p.y > def.ScreenHeight
}

func (p *PathMine) GetDeathEffect() def.DeathEffect {
	return def.DeathEffect{
		SpawnVisualEffect: func(cx, cy int, scene def.Scene) {
			if exp, err := effects.NewExplosionScaled(cx, cy, effects.ExplosionLarge, 3.0); err == nil {
				scene.Entities().Add(exp)
			}
		},
	}
}

func (p *PathMine) MarkAsDead(_ def.Scene) {
	p.dead = true
	p.frameCount = 0
	p.maxFrames = 60
}

func (p *PathMine) IsDead() bool {
	return p.dead
}

func (p *PathMine) TakeDamage(amount int) {
	if p.dead {
		return
	}
	p.hp -= amount
	if p.hp <= 0 {
		p.hp = 0
		p.dead = true
	}
}

func (p *PathMine) CurrentHP() int {
	return p.hp
}

func (p *PathMine) MaxHP() int {
	return p.maxHP
}

func (p *PathMine) BlastRadius() float64 {
	return pathMineBlastRadius
}

func (p *PathMine) BlastDamage() int {
	return pathMineBlastDamage
}

func (p *PathMine) ScoreValue() int {
	return pathMineValue
}

// ─── PathRangeMine (path-following + proximity countdown) ─────────────────────
// PathRangeMine combines PathMine path-following with RangeMine proximity
// detonation. Its violet lights flash rapidly once the countdown is active.
// It detonates after the player stays within its detection radius for 10
// consecutive seconds, or on direct contact.

const (
	pathRangeMineDetectionRadius = 120.0
	pathRangeMineDetonateFrames  = 600 // 10 seconds at 60 fps
	pathRangeMineBlastRadius     = 72.0
	pathRangeMineBlastDamage     = 999
	pathRangeMineMaxHP           = 1
	pathRangeMineValue           = 75
	pathRangeMineDrift           = 0.8
)

type PathRangeMine struct {
	x, y            int
	fx, fy          float64
	drift           float64
	width, height   int
	idleSprite      *draw.ColorMatrix
	activeSprite    *draw.ColorMatrix
	cachedImg       *ebit.Image
	pixelBuf        []byte
	path            []PathSegment
	pathFrame       int
	segmentTick     int
	proximityFrames int
	active          bool
	dead            bool
	frameCount      int
	maxFrames       int
	hp, maxHP       int
}

func NewPathRangeMine(x, y int, path []PathSegment) (*PathRangeMine, error) {
	idle, err := loadMineSprite("path_range_mine_idle.yaml")
	if err != nil {
		return nil, err
	}
	active, err := loadMineSprite("path_range_mine_active.yaml")
	if err != nil {
		return nil, err
	}
	w, h := idle.Dimensions()
	scaledW := int(float64(w) * enemyDrawScale)
	scaledH := int(float64(h) * enemyDrawScale)
	startX := x - scaledW/2
	return &PathRangeMine{
		x:            startX,
		y:            y,
		fx:           float64(startX),
		fy:           float64(y),
		drift:        pathRangeMineDrift,
		width:        scaledW,
		height:       scaledH,
		idleSprite:   idle,
		activeSprite: active,
		cachedImg:    ebit.NewImage(w, h),
		pixelBuf:     make([]byte, w*h*4),
		path:         path,
		hp:           pathRangeMineMaxHP,
		maxHP:        pathRangeMineMaxHP,
	}, nil
}

func (p *PathRangeMine) Type() def.EntityType {
	return def.EntityTypeEnemy
}

func (p *PathRangeMine) Location() (int, int) {
	return p.x, p.y
}

func (p *PathRangeMine) Dimensions() (int, int) {
	return p.width, p.height
}

func (p *PathRangeMine) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(p.x+p.width < ox || p.x > ox+ow || p.y+p.height < oy || p.y > oy+oh)
}

func (p *PathRangeMine) Act(scene def.Scene) {
	if p.dead {
		p.frameCount++
		return
	}
	p.fy += p.drift
	if len(p.path) > 0 {
		seg := p.path[p.pathFrame]
		p.fx += seg.VX
		p.fy += seg.VY
		p.segmentTick++
		if p.segmentTick >= seg.Frames {
			p.segmentTick = 0
			p.pathFrame = (p.pathFrame + 1) % len(p.path)
		}
	}
	if p.fx < 0 {
		p.fx = 0
	}
	if p.fx+float64(p.width) > float64(scene.Width()) {
		p.fx = float64(scene.Width() - p.width)
	}
	p.x = int(p.fx)
	p.y = int(p.fy)
	players := scene.Entities().Get(def.EntityTypePlayer)
	if len(players) == 0 {
		p.proximityFrames = 0
		p.active = false
		return
	}
	px, py := players[0].Location()
	pw, ph := players[0].Dimensions()
	mcx := p.fx + float64(p.width)/2
	mcy := p.fy + float64(p.height)/2
	dx := float64(px+pw/2) - mcx
	dy := float64(py+ph/2) - mcy
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist <= pathRangeMineDetectionRadius {
		p.proximityFrames++
		p.active = true
	} else {
		p.proximityFrames = 0
		p.active = false
	}
}

func (p *PathRangeMine) ReadyToDetonate() bool {
	return p.proximityFrames >= pathRangeMineDetonateFrames
}

func (p *PathRangeMine) SetDrift(drift float64) {
	p.drift = drift
}

func (p *PathRangeMine) GetDrift() float64 {
	return p.drift
}

func (p *PathRangeMine) ResetDrift() {
	p.drift = pathRangeMineDrift
}

func (p *PathRangeMine) Draw(img *ebit.Image) {
	sprite := p.idleSprite
	if p.active {
		sprite = p.activeSprite
	}
	draw.DrawScaled(img, p.cachedImg, p.pixelBuf, sprite, float64(p.x), float64(p.y), enemyDrawScale)
}

func (p *PathRangeMine) CanBeRemoved() bool {
	if p.dead {
		return p.frameCount >= p.maxFrames
	}
	return p.y > def.ScreenHeight
}

func (p *PathRangeMine) GetDeathEffect() def.DeathEffect {
	return def.DeathEffect{
		SpawnVisualEffect: func(cx, cy int, scene def.Scene) {
			if exp, err := effects.NewExplosionScaled(cx, cy, effects.ExplosionLarge, 3.0); err == nil {
				scene.Entities().Add(exp)
			}
		},
	}
}

func (p *PathRangeMine) MarkAsDead(_ def.Scene) {
	p.dead = true
	p.frameCount = 0
	p.maxFrames = 60
}

func (p *PathRangeMine) IsDead() bool {
	return p.dead
}

func (p *PathRangeMine) TakeDamage(amount int) {
	if p.dead {
		return
	}
	p.hp -= amount
	if p.hp <= 0 {
		p.hp = 0
		p.dead = true
	}
}

func (p *PathRangeMine) CurrentHP() int {
	return p.hp
}

func (p *PathRangeMine) MaxHP() int {
	return p.maxHP
}

func (p *PathRangeMine) BlastRadius() float64 {
	return pathRangeMineBlastRadius
}

func (p *PathRangeMine) BlastDamage() int {
	return pathRangeMineBlastDamage
}

func (p *PathRangeMine) ScoreValue() int {
	return pathRangeMineValue
}
