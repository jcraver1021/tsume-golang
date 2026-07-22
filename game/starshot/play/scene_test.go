package play_test

import (
	"testing"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/play"
	"tsumegolang/game/starshot/testutil"
)

var _ def.Scene = (*play.Scene)(nil)

func TestSceneDimensions(t *testing.T) {
	state := &play.GameState{Mode: play.GameModeIntro, Wave: 1}
	scene := play.NewScene(state)

	gotWidth := scene.Width()
	gotHeight := scene.Height()

	if gotWidth != def.ScreenWidth {
		t.Errorf("Width() = %d, want %d", gotWidth, def.ScreenWidth)
	}

	if gotHeight != def.ScreenHeight {
		t.Errorf("Height() = %d, want %d", gotHeight, def.ScreenHeight)
	}
}

func TestSceneEntities(t *testing.T) {
	state := &play.GameState{Mode: play.GameModeIntro, Wave: 1}
	scene := play.NewScene(state)

	if scene.Entities() == nil {
		t.Fatal("Entities() = nil, want non-nil EntityCollection")
	}
}

func TestSceneUpdate(t *testing.T) {
	state := &play.GameState{Mode: play.GameModeIntro, Wave: 1}
	scene := play.NewScene(state)

	actCount := 0
	trackingEntity := &actTrackingEntity{
		MockEntity: testutil.NewMockEntity(def.EntityTypePlayer),
		onAct: func() {
			actCount++
		},
	}

	scene.Entities().Add(trackingEntity)
	scene.Update()

	if actCount != 1 {
		t.Errorf("Act() called %d times after first Update, want 1", actCount)
	}

	scene.Update()

	if actCount != 2 {
		t.Errorf("Act() called %d times after second Update, want 2", actCount)
	}
}

func TestSceneIntroModeHasEntities(t *testing.T) {
	state := &play.GameState{Mode: play.GameModeIntro, Wave: 1}
	scene := play.NewScene(state)

	count := 0
	for range scene.Entities().IterateForUpdate() {
		count++
	}

	if count == 0 {
		t.Error("intro mode entity count = 0, want > 0 (should seed initial entities)")
	}
}

// actTrackingEntity wraps an entity and tracks Act calls.
type actTrackingEntity struct {
	*testutil.MockEntity
	onAct func()
}

func (a *actTrackingEntity) Act(s def.Scene) {
	if a.onAct != nil {
		a.onAct()
	}
}
