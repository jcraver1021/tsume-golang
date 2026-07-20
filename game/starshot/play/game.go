package play

import (
	"image/color"
	"math"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
	"tsumegolang/game/starshot/entity/background"
	"tsumegolang/game/starshot/entity/effects"
	"tsumegolang/game/starshot/entity/player"
	"tsumegolang/game/starshot/entity/projectile"
)

const (
	GameTitle = "Starshot"
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
	SlowdownActive     bool
	SlowdownMultiplier float64 // 1.0 = normal, 0.3 = 30% speed
	SlowdownFramesLeft int
	PlayerDied         bool // Tracks if player died this wave
}

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
					g.handleDeath(mortal)
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
				if d, ok := p.(def.Damageable); ok {
					g.applyDamage(p, d.MaxHP())
				}
				if mortal, ok := p.(def.Mortal); ok && mortal.IsDead() {
					g.handleDeath(mortal)
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
			// Chaser takes 1 damage per frame it overlaps an asteroid
			if d, ok := enemy.(def.Damageable); ok {
				d.TakeDamage(1)
				if isMortal && mortalEnemy.IsDead() {
					g.handleDeath(mortalEnemy)
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
			g.applyBombBlast(bomb)
			g.handleDeath(bomb)
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
			g.applyBombBlast(bomb)
			g.handleDeath(bomb)
			break
		}
	}
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
		g.handleDeath(mortal)
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
			if mortal, ok := target.(def.Mortal); ok && mortal.IsDead() {
				g.handleDeath(mortal)
			}
		}
	}
}

// handleDeath spawns explosion and activates slowdown for mortal entities
func (g *Game) handleDeath(mortal def.Mortal) {
	// Mark entity as dead
	mortal.MarkAsDead(g.Scene)

	// Track if it was the player that died
	if mortal.Type() == def.EntityTypePlayer {
		g.State.PlayerDied = true
	}

	// Get death effect specification
	deathEffect := mortal.GetDeathEffect()

	// Player handles its own explosion via sprite composition
	if mortal.Type() == def.EntityTypePlayer {
		// Load explosion sprite from effects package
		explosionSprite, err := effects.LoadExplosionSprite(deathEffect.ExplosionSize)
		if err == nil {
			// Compose explosion onto player sprite
			if playerEntity, ok := mortal.(interface{ ComposeExplosion(*draw.ColorMatrix) error }); ok {
				playerEntity.ComposeExplosion(explosionSprite)
			}
		}
	} else {
		// For other entities, spawn a separate explosion entity
		ex, ey := mortal.Location()
		ew, eh := mortal.Dimensions()

		explosion, err := effects.NewExplosion(
			ex+ew/2, // Center on entity
			ey+eh/2,
			g.Scene,
			deathEffect.ExplosionSize,
		)
		if err != nil {
			// Log error but don't crash the game
			return
		}

		g.Scene.Entities().Add(explosion)
	}

	// Activate slowdown if specified
	if deathEffect.SlowdownDuration > 0 {
		g.State.ActivateSlowdown(
			deathEffect.SlowdownMultiplier,
			deathEffect.SlowdownDuration,
		)
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
