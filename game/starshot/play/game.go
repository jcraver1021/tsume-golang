package play

import (
	ebit "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"tsumegolang/game/starshot/def"
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
	Mode GameMode
	Wave int
}

func NewGameState() *GameState {
	return &GameState{
		Mode: GameModeIntro,
		Wave: 1,
	}
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
	}

	return nil
}

// checkCollisions checks for player-obstacle collisions
func (g *Game) checkCollisions() {
	players := g.Scene.Entities().Get(def.EntityTypePlayer)
	obstacles := g.Scene.Entities().Get(def.EntityTypeObstacle)

	for _, p := range players {
		for _, obstacle := range obstacles {
			if def.Collides(p, obstacle) {
				// Player hit an obstacle - game over
				g.State.Mode = GameModeGameOver
				g.Scene = NewScene(g.State)
				return
			}
		}
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
