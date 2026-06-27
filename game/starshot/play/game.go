package play

import (
	"image/color"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/draw"
	"tsumegolang/game/starshot/entity/background"
	"tsumegolang/game/starshot/entity/effects"
	"tsumegolang/game/starshot/entity/player"
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

// checkCollisions checks for player-obstacle collisions
func (g *Game) checkCollisions() {
	players := g.Scene.Entities().Get(def.EntityTypePlayer)
	obstacles := g.Scene.Entities().Get(def.EntityTypeObstacle)

	for _, p := range players {
		// Skip if already dead
		if mortal, ok := p.(def.Mortal); ok && mortal.IsDead() {
			continue
		}

		for _, obstacle := range obstacles {
			if def.Collides(p, obstacle) {
				// Trigger death effect if entity is mortal
				if mortal, ok := p.(def.Mortal); ok {
					g.handleDeath(mortal)
				}
				return
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
			MoveUp:    ebit.IsKeyPressed(ebit.KeyArrowUp),
			MoveDown:  ebit.IsKeyPressed(ebit.KeyArrowDown),
			MoveLeft:  ebit.IsKeyPressed(ebit.KeyArrowLeft),
			MoveRight: ebit.IsKeyPressed(ebit.KeyArrowRight),
			Shoot:     ebit.IsKeyPressed(ebit.KeySpace),
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
