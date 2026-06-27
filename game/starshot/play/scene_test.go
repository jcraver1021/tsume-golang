package play_test

import (
	"testing"

	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/play"
	"tsumegolang/game/starshot/testutil"
)

func TestSceneImplementsInterface(t *testing.T) {
	state := &play.GameState{Mode: play.GameModeIntro, Wave: 1}
	scene := play.NewScene(state)
	var _ def.Scene = scene
}

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

	entities := scene.Entities()

	if entities == nil {
		t.Fatal("Entities() = nil, want non-nil EntityCollection")
	}

	// Verify it's a working collection
	var _ def.EntityCollection = entities
}

func TestSceneUpdate(t *testing.T) {
	state := &play.GameState{Mode: play.GameModeIntro, Wave: 1}
	scene := play.NewScene(state)

	// Add trackable entities
	actCount := 0
	trackingEntity := &actTrackingEntity{
		MockEntity: testutil.NewMockEntity(def.EntityTypePlayer),
		onAct: func() {
			actCount++
		},
	}

	scene.Entities().Add(trackingEntity)

	// Call Update
	scene.Update()

	got := actCount
	want := 1

	if got != want {
		t.Errorf("Act() called %d times, want %d", got, want)
	}

	// Update again to verify multiple calls work
	scene.Update()

	got = actCount
	want = 2

	if got != want {
		t.Errorf("after second Update, Act() called %d times total, want %d", got, want)
	}
}

func TestSceneIntroModeHasEntities(t *testing.T) {
	state := &play.GameState{Mode: play.GameModeIntro, Wave: 1}
	scene := play.NewScene(state)

	// Intro mode should seed initial entities
	count := 0
	for range scene.Entities().IterateForUpdate() {
		count++
	}

	if count == 0 {
		t.Error("intro mode entity count = 0, want > 0 (should seed initial entities)")
	}

	t.Logf("Intro mode initialized with %d entities", count)
}

// actTrackingEntity wraps an entity and tracks Act calls
type actTrackingEntity struct {
	*testutil.MockEntity
	onAct func()
}

func (a *actTrackingEntity) Act(s def.Scene) {
	if a.onAct != nil {
		a.onAct()
	}
}
