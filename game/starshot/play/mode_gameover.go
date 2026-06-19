package play

import (
	"image/color"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/background"
	"tsumegolang/game/starshot/entity/environment"
)

func initGameOverMode(b def.Scene) {
	// Starfield background (same as intro)
	b.Entities().Add(environment.NewSpace(introStarDensity, b))

	// Game Over banner - centered at upper third
	gameOver, err := background.NewBanner(
		"GAME OVER",
		def.ScreenWidth/2,
		def.ScreenHeight/3,
		36.0,
		color.RGBA{R: 255, G: 100, B: 100, A: 255}, // Reddish
	)
	if err == nil {
		b.Entities().Add(gameOver)
	}

	// Instruction banner - centered at lower third
	instruction, err := background.NewBanner(
		"Press SPACE to Continue",
		def.ScreenWidth/2,
		def.ScreenHeight*2/3,
		20.0,
		color.RGBA{R: 200, G: 200, B: 200, A: 255},
	)
	if err == nil {
		b.Entities().Add(instruction)
	}
}
