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
	tick     int // Global animation tick counter
	state    *GameState
}

// Interface implementation methods
func (s *Scene) Width() int                     { return s.width }
func (s *Scene) Height() int                    { return s.height }
func (s *Scene) Entities() def.EntityCollection { return s.entities }
func (s *Scene) Tick() int                      { return s.tick }

func NewScene(state *GameState) *Scene {
	scene := &Scene{
		entities: NewEntityStore(),
		width:    def.ScreenWidth,
		height:   def.ScreenHeight,
		state:    state,
	}

	switch state.Mode {
	case GameModeIntro:
		initIntroMode(scene)
	case GameModePlay:
		initPlayMode(scene, state)
	case GameModeGameOver:
		// Game over mode reuses the play scene with overlay banners
		// No initialization needed
	}

	return scene
}

func (s *Scene) Update() {
	s.tick++ // Increment global tick counter (always advance)

	// Check if we should skip this frame due to slowdown
	shouldUpdate := true
	if s.state.SlowdownActive {
		// Update every Nth frame based on multiplier
		// Example: 0.3x = update every 3rd frame (skip 2 out of 3)
		frameInterval := int(1.0 / s.state.SlowdownMultiplier)
		shouldUpdate = (s.tick % frameInterval) == 0

		// Decrement slowdown counter
		s.state.SlowdownFramesLeft--
		if s.state.SlowdownFramesLeft <= 0 {
			s.state.SlowdownActive = false
		}
	}

	if shouldUpdate {
		for e := range s.entities.IterateForUpdate() {
			e.Act(s)
		}
	}
}

func (s *Scene) Draw(screen *ebit.Image) {
	for e := range s.entities.IterateForDraw() {
		e.Draw(screen)
	}
}
