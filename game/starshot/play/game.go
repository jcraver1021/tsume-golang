package play

import (
	ebit "github.com/hajimehoshi/ebiten/v2"
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

	return nil
}

func (g *Game) handleInput() {
	if ebit.IsKeyPressed(g.controls.ExitKey) {
		g.exit = true
		return
	}

	// Handle mode-specific input
	switch g.State.Mode {
	case GameModeIntro:
		if ebit.IsKeyPressed(g.controls.StartKey) {
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
			if playerEntity, ok := p.(*player.Player); ok {
				playerEntity.SetPlayerAction(playerAction)
			}
		}
	case GameModePaused:
		// Handle paused input here
	case GameModeExitConfirm:
		// Handle exit confirmation input here
	case GameModeGameOver:
		// Handle game over input here
	}
}

func (g *Game) Draw(screen *ebit.Image) {
	g.Scene.Draw(screen)
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return def.ScreenWidth, def.ScreenHeight
}
