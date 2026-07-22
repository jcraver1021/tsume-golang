package play

import (
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/entity/player"
	"tsumegolang/game/starshot/entity/ui"
	"tsumegolang/game/starshot/entity/wave"
)

func initPlayMode(b def.Scene, state *GameState) {
	if len(b.Entities().Get(def.EntityTypePlayer)) == 0 {
		p, err := newPlayer(def.ScreenWidth/2, def.ScreenHeight-50)
		if err != nil {
			panic(err)
		}
		b.Entities().Add(p)
	}
	b.Entities().Add(ui.NewHUD(state))
	wave.LoadWave(b, state.Wave)
}

func newPlayer(x, y int) (*player.Player, error) {
	gun, err := player.NewBasicGun()
	if err != nil {
		return nil, err
	}
	launcher, err := player.NewBombLauncher()
	if err != nil {
		return nil, err
	}
	return player.NewPlayer(x, y, gun, launcher)
}
