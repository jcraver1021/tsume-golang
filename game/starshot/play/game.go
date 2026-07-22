package play

import (
	"image/color"
	"math"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/background"
	"tsumegolang/game/starshot/entity/player"
	"tsumegolang/game/starshot/entity/projectile"
)

const (
	GameTitle = "Starshot"
)

// KillMode controls which death effects fire when handleDeath is called.
// Use KillNormal for standard gameplay deaths.
// Use KillSilent when the killing mechanic should suppress all secondary effects
// (e.g. an ion beam that disables a mine without detonating it, or a black hole
// that consumes entities without triggering their explosions).
type KillMode int

const (
	KillNormal KillMode = iota // blast + visual effects fire as declared
	KillSilent                 // entity dies; no blast damage, no visual effect
	KillDisarm                 // entity dies; blast suppressed, visual effect fires
)

var (
	exitKey  = ebit.KeyEscape
	startKey = ebit.KeySpace
	pauseKey = ebit.KeyP
)

type GameControlSettings struct {
	ExitKey  ebit.Key
	StartKey ebit.Key
	PauseKey ebit.Key
}

func DefaultGameControlSettings() *GameControlSettings {
	return &GameControlSettings{
		ExitKey:  exitKey,
		StartKey: startKey,
		PauseKey: pauseKey,
	}
}

type GameState struct {
	Mode               GameMode
	Wave               int
	Score              int
	SlowdownActive     bool
	SlowdownMultiplier float64 // 1.0 = normal, 0.3 = 30% speed
	SlowdownFramesLeft int
	PlayerDied         bool // Tracks if player died this wave
}

func (s *GameState) GetWave() int  { return s.Wave }
func (s *GameState) GetScore() int { return s.Score }

func NewGameState() *GameState {
	return &GameState{
		Mode:               GameModeIntro,
		Wave:               1,
		SlowdownActive:     false,
		SlowdownMultiplier: 1.0,
		SlowdownFramesLeft: 0,
		PlayerDied:         false,
	}
}

// Reset reinitializes the game state to start a new game.
func (gs *GameState) Reset() {
	gs.Mode = GameModeIntro
	gs.Wave = 1
	gs.Score = 0
	gs.SlowdownActive = false
	gs.SlowdownMultiplier = 1.0
	gs.SlowdownFramesLeft = 0
	gs.PlayerDied = false
}

// ActivateSlowdown initiates game slowdown for dramatic effect
func (gs *GameState) ActivateSlowdown(multiplier float64, frames int) {
	if multiplier <= 0 || multiplier > 1.0 || frames <= 0 {
		return // Skip invalid slowdown requests
	}
	gs.SlowdownActive = true
	gs.SlowdownMultiplier = multiplier
	gs.SlowdownFramesLeft = frames
}

type Game struct {
	Scene    *Scene
	State    *GameState
	controls *GameControlSettings

	exit   bool
	paused bool
}

func NewGame() *Game {
	state := NewGameState()

	return &Game{
		Scene:    NewScene(state),
		State:    state,
		controls: DefaultGameControlSettings(),
	}
}

func (g *Game) Update() error {
	g.handleInput()
	if g.exit {
		return ebit.Termination
	}

	g.Scene.Update()

	// Check collisions in play mode
	if g.State.Mode == GameModePlay {
		g.checkCollisions()

		// Check if player death animation is complete
		if g.State.PlayerDied && !g.State.SlowdownActive {
			// Player is dead and slowdown expired → transition to game over
			g.State.Mode = GameModeGameOver
			// Add game over banners to existing scene instead of recreating
			g.addGameOverBanners()
			return nil
		}
	}

	return nil
}

const bulletDamage = 1

// checkCollisions checks for all relevant entity collisions
func (g *Game) checkCollisions() {
	players := g.Scene.Entities().Get(def.EntityTypePlayer)
	obstacles := g.Scene.Entities().Get(def.EntityTypeObstacle)
	bullets := g.Scene.Entities().Get(def.EntityTypeTeam)
	enemies := g.Scene.Entities().Get(def.EntityTypeEnemy)

	// Player vs obstacles
	for _, p := range players {
		if mortal, ok := p.(def.Mortal); ok && mortal.IsDead() {
			continue
		}
		for _, obs := range obstacles {
			if mortal, ok := obs.(def.Mortal); ok && mortal.IsDead() {
				continue
			}
			if def.Collides(p, obs) {
				if d, ok := p.(def.Damageable); ok {
					g.applyDamage(p, d.MaxHP())
				}
				if mortal, ok := p.(def.Mortal); ok && mortal.IsDead() {
					g.handleDeath(mortal, KillNormal)
					return
				}
			}
		}
	}

	// Player vs enemies
	for _, p := range players {
		if mortal, ok := p.(def.Mortal); ok && mortal.IsDead() {
			continue
		}
		for _, enemy := range enemies {
			if mortal, ok := enemy.(def.Mortal); ok && mortal.IsDead() {
				continue
			}
			if def.Collides(p, enemy) {
				// Explosive enemies (e.g. mines) detonate on player contact
				if mortalEnemy, ok := enemy.(def.Mortal); ok && !mortalEnemy.IsDead() {
					if _, ok := enemy.(def.Explosive); ok {
						g.handleDeath(mortalEnemy, KillNormal)
					}
				}
				if d, ok := p.(def.Damageable); ok {
					g.applyDamage(p, d.MaxHP())
				}
				if mortal, ok := p.(def.Mortal); ok && mortal.IsDead() {
					g.handleDeath(mortal, KillNormal)
					return
				}
			}
		}
	}

	// Enemies vs obstacles
	for _, enemy := range enemies {
		mortalEnemy, isMortal := enemy.(def.Mortal)
		if isMortal && mortalEnemy.IsDead() {
			continue
		}
		for _, obs := range obstacles {
			if mortal, ok := obs.(def.Mortal); ok && mortal.IsDead() {
				continue
			}
			if !def.Collides(enemy, obs) {
				continue
			}
			// Explosive enemies (e.g. mines) detonate on obstacle contact
			if _, ok := enemy.(def.Explosive); ok {
				if isMortal && !mortalEnemy.IsDead() {
					g.handleDeath(mortalEnemy, KillNormal)
				}
				break
			}
			// Non-explosive enemies (chasers) take 1 damage per frame of overlap
			if d, ok := enemy.(def.Damageable); ok {
				d.TakeDamage(1)
				if isMortal && mortalEnemy.IsDead() {
					g.handleDeath(mortalEnemy, KillNormal)
				}
			}
			break
		}
	}

	// Bullets vs obstacles and enemies
	for _, b := range bullets {
		bullet, ok := b.(*projectile.Bullet)
		if !ok || bullet.CanBeRemoved() {
			continue
		}

		bx, by := bullet.Location()
		bw, bh := bullet.Dimensions()
		impactX := float64(bx + bw/2)
		_ = by
		_ = bh

		for _, obs := range obstacles {
			if mortal, ok := obs.(def.Mortal); ok && mortal.IsDead() {
				continue
			}
			if !def.Collides(b, obs) {
				continue
			}
			bullet.MarkDestroyed()
			g.applyBulletHit(obs, impactX)
			break
		}

		if bullet.CanBeRemoved() {
			continue
		}

		for _, enemy := range enemies {
			if mortal, ok := enemy.(def.Mortal); ok && mortal.IsDead() {
				continue
			}
			if !def.Collides(b, enemy) {
				continue
			}
			bullet.MarkDestroyed()
			g.applyBulletHit(enemy, impactX)
			break
		}
	}

	// Enemy bullets vs player, obstacles, and enemies
	for _, eb := range g.Scene.Entities().Get(def.EntityTypeEnemyTeam) {
		bullet, ok := eb.(*projectile.EnemyBullet)
		if !ok || bullet.CanBeRemoved() {
			continue
		}

		for _, p := range players {
			if mortal, ok := p.(def.Mortal); ok && mortal.IsDead() {
				continue
			}
			if !def.Collides(eb, p) {
				continue
			}
			bullet.MarkDestroyed()
			g.applyDamage(p, projectile.EnemyBulletDamage)
			if mortal, ok := p.(def.Mortal); ok && mortal.IsDead() {
				g.handleDeath(mortal, KillNormal)
				return
			}
			break
		}

		if bullet.CanBeRemoved() {
			continue
		}

		for _, obs := range obstacles {
			if mortal, ok := obs.(def.Mortal); ok && mortal.IsDead() {
				continue
			}
			if !def.Collides(eb, obs) {
				continue
			}
			bullet.MarkDestroyed()
			g.applyBulletHit(obs, 0)
			break
		}

		if bullet.CanBeRemoved() {
			continue
		}

		for _, enemy := range enemies {
			if mortal, ok := enemy.(def.Mortal); ok && mortal.IsDead() {
				continue
			}
			if !def.Collides(eb, enemy) {
				continue
			}
			bullet.MarkDestroyed()
			g.applyBulletHit(enemy, 0)
			break
		}
	}

	// Self-detonating enemies (e.g. proximity range mines)
	for _, e := range enemies {
		if mortal, ok := e.(def.Mortal); !ok || mortal.IsDead() {
			continue
		}
		if sd, ok := e.(def.SelfDetonating); ok && sd.ReadyToDetonate() {
			g.handleDeath(e.(def.Mortal), KillNormal)
		}
	}

	// Bombs vs obstacles and enemies
	for _, b := range bullets {
		bomb, ok := b.(*projectile.Bomb)
		if !ok || bomb.CanBeRemoved() {
			continue
		}

		detonated := false
		for _, obs := range obstacles {
			if m, ok := obs.(def.Mortal); ok && m.IsDead() {
				continue
			}
			if !def.Collides(b, obs) {
				continue
			}
			g.handleDeath(bomb, KillNormal)
			detonated = true
			break
		}

		if detonated {
			continue
		}

		for _, enemy := range enemies {
			if m, ok := enemy.(def.Mortal); ok && m.IsDead() {
				continue
			}
			if !def.Collides(b, enemy) {
				continue
			}
			g.handleDeath(bomb, KillNormal)
			break
		}
	}
}

// isWaveCleared returns true when every enemy entity is dead.
// Returns false when there are no enemies (wave not yet started or already removed).
func (g *Game) isWaveCleared() bool {
	enemies := g.Scene.Entities().Get(def.EntityTypeEnemy)
	if len(enemies) == 0 {
		return false
	}
	for _, e := range enemies {
		if mortal, ok := e.(def.Mortal); !ok || !mortal.IsDead() {
			return false
		}
	}
	return true
}

// applyDamage deals damage to an entity via the Damageable interface.
func (g *Game) applyDamage(e def.Entity, amount int) {
	if d, ok := e.(def.Damageable); ok {
		d.TakeDamage(amount)
	}
}

// applyBulletHit deals bullet damage to a target, applies impulse if it survives,
// and triggers death handling if HP reaches zero.
func (g *Game) applyBulletHit(target def.Entity, impactX float64) {
	damageable, isDamageable := target.(def.Damageable)
	if !isDamageable {
		return
	}

	damageable.TakeDamage(bulletDamage)

	mortal, isMortal := target.(def.Mortal)
	if isMortal && mortal.IsDead() {
		g.handleDeath(mortal, KillNormal)
		return
	}

	// Target survived — apply impulse proportional to hit offset from center
	if imp, ok := target.(def.Impulsable); ok {
		tx, _ := target.Location()
		tw, _ := target.Dimensions()
		centerX := float64(tx + tw/2)
		halfW := float64(tw) / 2
		if halfW < 1 {
			halfW = 1
		}
		offsetX := impactX - centerX
		// Small upward push plus lateral component based on where bullet hit
		imp.ApplyImpulse(-offsetX/halfW*12.0, -2.0)
	}
}

// applyBombBlast damages every Damageable entity whose center lies within the
// bomb's blast radius, triggering handleDeath on any that are killed.
func (g *Game) applyBombBlast(exp def.Explosive) {
	bx, by := exp.Location()
	bw, bh := exp.Dimensions()
	cx := float64(bx + bw/2)
	cy := float64(by + bh/2)
	radius := exp.BlastRadius()

	targets := append(
		g.Scene.Entities().Get(def.EntityTypeObstacle),
		g.Scene.Entities().Get(def.EntityTypeEnemy)...,
	)
	targets = append(targets, g.Scene.Entities().Get(def.EntityTypePlayer)...)

	for _, target := range targets {
		if m, ok := target.(def.Mortal); ok && m.IsDead() {
			continue
		}
		tx, ty := target.Location()
		tw, th := target.Dimensions()
		tcx := float64(tx + tw/2)
		tcy := float64(ty + th/2)
		dx := tcx - cx
		dy := tcy - cy
		if math.Sqrt(dx*dx+dy*dy) > radius {
			continue
		}
		if d, ok := target.(def.Damageable); ok {
			d.TakeDamage(exp.BlastDamage())
			// Chain reaction: if target is also Explosive, handleDeath fires its blast too
			if mortal, ok := target.(def.Mortal); ok && mortal.IsDead() {
				g.handleDeath(mortal, KillNormal)
			}
		}
	}
}

// handleDeath marks an entity dead, awards points, and — unless mode is
// KillSilent — fires blast damage and visual effects. Entities that implement
// def.Explosive have their blast applied automatically, so new explosive types
// require no changes to checkCollisions; just implement the interface.
func (g *Game) handleDeath(mortal def.Mortal, mode KillMode) {
	mortal.MarkAsDead(g.Scene)

	if mortal.Type() == def.EntityTypePlayer {
		g.State.PlayerDied = true
	}

	if scorer, ok := mortal.(def.Scorer); ok {
		g.State.Score += scorer.ScoreValue()
	}

	if mode == KillSilent {
		return
	}

	if mode != KillDisarm {
		if exp, ok := mortal.(def.Explosive); ok {
			g.applyBombBlast(exp)
		}
	}

	deathEffect := mortal.GetDeathEffect()
	ex, ey := mortal.Location()
	ew, eh := mortal.Dimensions()

	if deathEffect.SpawnVisualEffect != nil {
		deathEffect.SpawnVisualEffect(ex+ew/2, ey+eh/2, g.Scene)
	}

	if deathEffect.SlowdownDuration > 0 {
		g.State.ActivateSlowdown(deathEffect.SlowdownMultiplier, deathEffect.SlowdownDuration)
	}
}

// addGameOverBanners adds game over UI elements to the existing scene
func (g *Game) addGameOverBanners() {
	// Game Over banner - centered at upper third
	gameOver, err := background.NewUIBanner(
		"GAME OVER",
		def.ScreenWidth/2,
		def.ScreenHeight/3,
		36.0,
		color.RGBA{R: 255, G: 100, B: 100, A: 255}, // Reddish
	)
	if err == nil {
		g.Scene.Entities().Add(gameOver)
	}

	// Instruction banner - centered at lower third
	instruction, err := background.NewUIBanner(
		"Press SPACE to Continue",
		def.ScreenWidth/2,
		def.ScreenHeight*2/3,
		20.0,
		color.RGBA{R: 200, G: 200, B: 200, A: 255},
	)
	if err == nil {
		g.Scene.Entities().Add(instruction)
	}
}

func (g *Game) handleInput() {
	if ebit.IsKeyPressed(g.controls.ExitKey) {
		g.exit = true
		return
	}

	// Handle mode-specific input
	switch g.State.Mode {
	case GameModeIntro:
		if inpututil.IsKeyJustPressed(g.controls.StartKey) {
			g.State.Mode = GameModePlay
			g.Scene = NewScene(g.State)
		}
	case GameModePlay:
		playerAction := player.PlayerAction{
			MoveUp:         ebit.IsKeyPressed(ebit.KeyArrowUp),
			MoveDown:       ebit.IsKeyPressed(ebit.KeyArrowDown),
			MoveLeft:       ebit.IsKeyPressed(ebit.KeyArrowLeft),
			MoveRight:      ebit.IsKeyPressed(ebit.KeyArrowRight),
			ShootPrimary:   ebit.IsKeyPressed(ebit.KeySpace),
			ShootSecondary: ebit.IsKeyPressed(ebit.KeyZ),
		}

		// Apply player action to all player entities (even zero or more than one)
		players := g.Scene.Entities().Get(def.EntityTypePlayer)
		for _, p := range players {
			if playerEntity, ok := p.(player.PlayerController); ok {
				playerEntity.SetPlayerAction(playerAction)
			}
		}
	case GameModePaused:
		// Handle paused input here
	case GameModeExitConfirm:
		// Handle exit confirmation input here
	case GameModeGameOver:
		if inpututil.IsKeyJustPressed(g.controls.StartKey) {
			// Return to intro
			g.State.Mode = GameModeIntro
			g.State.Wave = 1
			g.State.PlayerDied = false
			g.Scene = NewScene(g.State)
		}
	}
}

func (g *Game) Draw(screen *ebit.Image) {
	g.Scene.Draw(screen)
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return def.ScreenWidth, def.ScreenHeight
}
