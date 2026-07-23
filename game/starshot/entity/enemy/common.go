package enemy

import (
	"math"

	"tsumegolang/game/starshot/def"
)

// enemyHealth is embedded in every enemy type to share HP/death-animation state.
type enemyHealth struct {
	hp, maxHP             int
	dead                  bool
	frameCount, maxFrames int
}

func (e *enemyHealth) TakeDamage(amount int) {
	if e.dead {
		return
	}
	e.hp -= amount
	if e.hp <= 0 {
		e.hp = 0
		e.dead = true
	}
}

func (e *enemyHealth) startDeath(maxFrames int) {
	e.dead = true
	e.frameCount = 0
	e.maxFrames = maxFrames
}

func (e *enemyHealth) tickDeath() {
	e.frameCount++
}

func (e *enemyHealth) deathComplete() bool {
	return e.dead && e.frameCount >= e.maxFrames
}

func (e *enemyHealth) CurrentHP() int {
	return e.hp
}

func (e *enemyHealth) MaxHP() int {
	return e.maxHP
}

func (e *enemyHealth) IsDead() bool {
	return e.dead
}

// aabbOverlaps reports whether the AABB (ax, ay, aw, ah) overlaps other's bounding box.
func aabbOverlaps(ax, ay, aw, ah int, other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(ax+aw < ox || ax > ox+ow || ay+ah < oy || ay > oy+oh)
}

// perceiveFighter builds a Perception for fighter enemies (Chaser, Hunter).
// It emits SignalSelf, SignalPlayer, and SignalObstacle for hazards within lookahead pixels.
func perceiveFighter(hp, maxHP, lookahead int, fx, fy float64, w, h int, scene def.Scene) def.Perception {
	p := def.Perception{
		{Kind: def.SignalSelf, Condition: def.ConditionFor(hp, maxHP)},
	}
	players := scene.Entities().Get(def.EntityTypePlayer)
	if len(players) == 0 {
		return p
	}
	px, py := players[0].Location()
	pw, ph := players[0].Dimensions()
	myCX := fx + float64(w)/2
	myCY := fy + float64(h)/2
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
		if along < 0 || along > float64(lookahead) {
			continue
		}
		perpX := obsCX - along*nx
		perpY := obsCY - along*ny
		if math.Sqrt(perpX*perpX+perpY*perpY) >= float64(ow/2+w/2)*1.4 {
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
