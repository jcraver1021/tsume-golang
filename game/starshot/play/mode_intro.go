package play

import (
	"image/color"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/background"
	"tsumegolang/game/starshot/entity/environment"
)

const (
	introStarDensity = 0.05
)

func initIntroMode(b def.Scene) {
	// Starfield background
	b.Entities().Add(environment.NewSpace(introStarDensity, b))

	// Title banner - centered at upper third
	title, err := background.NewBanner(
		"STARSHOT",
		def.ScreenWidth/2,
		def.ScreenHeight/3,
		48.0,
		color.RGBA{R: 255, G: 255, B: 255, A: 255},
	)
	if err == nil {
		b.Entities().Add(title)
	}

	// Instruction banner - centered at lower third
	instruction, err := background.NewBanner(
		"Press SPACE to Start",
		def.ScreenWidth/2,
		def.ScreenHeight*2/3,
		24.0,
		color.RGBA{R: 200, G: 200, B: 200, A: 255},
	)
	if err == nil {
		b.Entities().Add(instruction)
	}
}
