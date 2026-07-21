package enemy

// drifters.go — passive enemies that descend without actively targeting the player.
// Add new drift-style enemies here; each section covers one type.

import (
	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
	"tsumegolang/game/starshot/entity/effects"
)

// ─── Drifter (fast straight descent) ─────────────────────────────────────────
// Drifter falls straight down at high speed with no lateral movement or
// targeting. It is the simplest threat: dodge or shoot it before it passes.

const (
	drifterSpeed = 3.5
	drifterMaxHP = 1
	drifterValue = 5
)

type Drifter struct {
	x, y          int
	fx, fy        float64
	width, height int
	sprite        *draw.ColorMatrix
	cachedImg     *ebit.Image
	pixelBuf      []byte
	dead          bool
	frameCount    int
	maxFrames     int
	hp, maxHP     int
}

func NewDrifter(x, y int) (*Drifter, error) {
	data, err := spriteFiles.ReadFile("sprites/drifter.yaml")
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
	return &Drifter{
		x:         x - scaledW/2,
		y:         y,
		fx:        float64(x - scaledW/2),
		fy:        float64(y),
		width:     scaledW,
		height:    scaledH,
		sprite:    sprite,
		cachedImg: ebit.NewImage(w, h),
		pixelBuf:  make([]byte, w*h*4),
		hp:        drifterMaxHP,
		maxHP:     drifterMaxHP,
	}, nil
}

func (d *Drifter) Type() def.EntityType   { return def.EntityTypeEnemy }
func (d *Drifter) Location() (int, int)   { return d.x, d.y }
func (d *Drifter) Dimensions() (int, int) { return d.width, d.height }

func (d *Drifter) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(d.x+d.width < ox || d.x > ox+ow || d.y+d.height < oy || d.y > oy+oh)
}

func (d *Drifter) Act(_ def.Scene) {
	if d.dead {
		d.frameCount++
		return
	}
	d.fy += drifterSpeed
	d.y = int(d.fy)
}

func (d *Drifter) Draw(img *ebit.Image) {
	draw.DrawScaled(img, d.cachedImg, d.pixelBuf, d.sprite, float64(d.x), float64(d.y), enemyDrawScale)
}

func (d *Drifter) CanBeRemoved() bool {
	if d.dead {
		return d.frameCount >= d.maxFrames
	}
	return d.y > def.ScreenHeight
}

func (d *Drifter) GetDeathEffect() def.DeathEffect {
	return def.DeathEffect{
		SpawnVisualEffect: func(cx, cy int, scene def.Scene) {
			if exp, err := effects.NewExplosion(cx, cy, effects.ExplosionSmall); err == nil {
				scene.Entities().Add(exp)
			}
		},
	}
}

func (d *Drifter) MarkAsDead(_ def.Scene) { d.dead = true; d.frameCount = 0; d.maxFrames = 30 }
func (d *Drifter) IsDead() bool           { return d.dead }

func (d *Drifter) TakeDamage(amount int) {
	if d.dead {
		return
	}
	d.hp -= amount
	if d.hp <= 0 {
		d.hp = 0
		d.dead = true
	}
}

func (d *Drifter) CurrentHP() int  { return d.hp }
func (d *Drifter) MaxHP() int      { return d.maxHP }
func (d *Drifter) ScoreValue() int { return drifterValue }

// ─── Weaver (descends while steering around obstacles) ────────────────────────
// Weaver drifts downward and actively pushes sideways when an obstacle is
// directly ahead. The result is a weaving path through the asteroid field.
// It does not target the player but is hard to ignore in a dense field.

const (
	weaverDownSpeed    = 2.5
	weaverLookahead    = 100.0 // px below center to scan for obstacles
	weaverMaxVX        = 3.0
	weaverPushStrength = 0.5  // lateral force per obstacle per frame (scaled by proximity)
	weaverDecay        = 0.92 // vx multiplier per frame; returns to center after clearing obstacles
	weaverMaxHP        = 2
	weaverValue        = 15
)

type Weaver struct {
	x, y          int
	fx, fy        float64
	vx            float64
	width, height int
	sprite        *draw.ColorMatrix
	cachedImg     *ebit.Image
	pixelBuf      []byte
	dead          bool
	frameCount    int
	maxFrames     int
	hp, maxHP     int
}

func NewWeaver(x, y int) (*Weaver, error) {
	data, err := spriteFiles.ReadFile("sprites/weaver.yaml")
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
	return &Weaver{
		x:         x - scaledW/2,
		y:         y,
		fx:        float64(x - scaledW/2),
		fy:        float64(y),
		width:     scaledW,
		height:    scaledH,
		sprite:    sprite,
		cachedImg: ebit.NewImage(w, h),
		pixelBuf:  make([]byte, w*h*4),
		hp:        weaverMaxHP,
		maxHP:     weaverMaxHP,
	}, nil
}

func (w *Weaver) Type() def.EntityType   { return def.EntityTypeEnemy }
func (w *Weaver) Location() (int, int)   { return w.x, w.y }
func (w *Weaver) Dimensions() (int, int) { return w.width, w.height }

func (w *Weaver) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(w.x+w.width < ox || w.x > ox+ow || w.y+w.height < oy || w.y > oy+oh)
}

func (w *Weaver) Act(scene def.Scene) {
	if w.dead {
		w.frameCount++
		return
	}

	w.fy += weaverDownSpeed
	w.vx *= weaverDecay

	myCX := w.fx + float64(w.width)/2
	myCY := w.fy + float64(w.height)/2

	for _, obs := range scene.Entities().Get(def.EntityTypeObstacle) {
		ox, oy := obs.Location()
		ow, oh := obs.Dimensions()
		obsCX := float64(ox + ow/2)
		obsCY := float64(oy + oh/2)
		dy := obsCY - myCY
		if dy < 0 || dy > weaverLookahead {
			continue
		}
		// Closer obstacles produce a stronger lateral push.
		proximity := 1.0 - dy/weaverLookahead
		if obsCX > myCX {
			w.vx -= proximity * weaverPushStrength // obstacle to the right → push left
		} else {
			w.vx += proximity * weaverPushStrength // obstacle to the left → push right
		}
	}

	if w.vx > weaverMaxVX {
		w.vx = weaverMaxVX
	}
	if w.vx < -weaverMaxVX {
		w.vx = -weaverMaxVX
	}

	w.fx += w.vx
	if w.fx < 0 {
		w.fx = 0
		w.vx = 0
	}
	if w.fx+float64(w.width) > float64(scene.Width()) {
		w.fx = float64(scene.Width() - w.width)
		w.vx = 0
	}

	w.x = int(w.fx)
	w.y = int(w.fy)
}

func (w *Weaver) Draw(img *ebit.Image) {
	draw.DrawScaled(img, w.cachedImg, w.pixelBuf, w.sprite, float64(w.x), float64(w.y), enemyDrawScale)
}

func (w *Weaver) CanBeRemoved() bool {
	if w.dead {
		return w.frameCount >= w.maxFrames
	}
	return w.y > def.ScreenHeight
}

func (w *Weaver) GetDeathEffect() def.DeathEffect {
	return def.DeathEffect{
		SpawnVisualEffect: func(cx, cy int, scene def.Scene) {
			if exp, err := effects.NewExplosion(cx, cy, effects.ExplosionSmall); err == nil {
				scene.Entities().Add(exp)
			}
		},
	}
}

func (w *Weaver) MarkAsDead(_ def.Scene) { w.dead = true; w.frameCount = 0; w.maxFrames = 30 }
func (w *Weaver) IsDead() bool           { return w.dead }

func (w *Weaver) TakeDamage(amount int) {
	if w.dead {
		return
	}
	w.hp -= amount
	if w.hp <= 0 {
		w.hp = 0
		w.dead = true
	}
}

func (w *Weaver) CurrentHP() int  { return w.hp }
func (w *Weaver) MaxHP() int      { return w.maxHP }
func (w *Weaver) ScoreValue() int { return weaverValue }
