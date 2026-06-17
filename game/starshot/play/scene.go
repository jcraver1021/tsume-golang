package play

import (
	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
)

// Scene is the concrete implementation of def.Scene
type Scene struct {
	entities *EntityStore
	width    int
	height   int
}

// Interface implementation methods
func (s *Scene) Width() int                     { return s.width }
func (s *Scene) Height() int                    { return s.height }
func (s *Scene) Entities() def.EntityCollection { return s.entities }

func NewScene(state *GameState) *Scene {
	scene := &Scene{
		entities: NewEntityStore(),
		width:    def.ScreenWidth,
		height:   def.ScreenHeight,
	}

	switch state.Mode {
	case GameModeIntro:
		initIntroMode(scene)
	case GameModePlay:
		initPlayMode(scene, state)
	}

	return scene
}

func (s *Scene) Update() {
	for e := range s.entities.IterateForUpdate() {
		e.Act(s)
	}
}

func (s *Scene) Draw(screen *ebit.Image) {
	for e := range s.entities.IterateForDraw() {
		e.Draw(screen)
	}
}
