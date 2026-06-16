package play

import (
	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
)

const (
	GameTitle = "Starshot"
)

type Game struct {
	Scene *Scene
	mode  GameMode

	exit   bool
	paused bool
}

func NewGame() *Game {
	return &Game{
		Scene: NewScene(GameModeIntro),
		mode:  GameModeIntro,
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
	if ebit.IsKeyPressed(ebit.KeyEscape) {
		g.exit = true
		return
	}
}

func (g *Game) Draw(screen *ebit.Image) {
	g.Scene.Draw(screen)
}

func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return def.ScreenWidth, def.ScreenHeight
}
