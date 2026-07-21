package play_test

import (
	"testing"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/play"
	"tsumegolang/game/starshot/testutil"
)

// spyEntity records every def.Entity it sees when Act queries a given type.
// This reproduces the Chaser/Hunter pattern: one entity type reads another
// type's entities during Act (e.g. Enemy reads Obstacle to build Perception).
type spyEntity struct {
	*testutil.MockEntity
	queryType def.EntityType
	seen      []def.Entity
}

func newSpyEntity(t def.EntityType, queryType def.EntityType) *spyEntity {
	return &spyEntity{MockEntity: testutil.NewMockEntity(t), queryType: queryType}
}

func (s *spyEntity) Act(scene def.Scene) {
	s.seen = append(s.seen, scene.Entities().Get(s.queryType)...)
}

func (s *spyEntity) Draw(_ *ebit.Image) {}

func TestEntityStoreImplementsInterface(t *testing.T) {
	store := play.NewEntityStore()
	var _ def.EntityCollection = store
}

func TestEntityStoreAdd(t *testing.T) {
	store := play.NewEntityStore()
	entity := testutil.NewMockEntity(def.EntityTypePlayer)

	store.Add(entity)

	// Verify entity was added by iterating
	count := 0
	for range store.IterateForUpdate() {
		count++
	}

	if count != 1 {
		t.Errorf("entity count = %d, want 1", count)
	}
}

func TestEntityStoreIterateForUpdateOrder(t *testing.T) {
	store := play.NewEntityStore()

	// Add entities in random order
	background := testutil.NewMockEntity(def.EntityTypeBackground)
	player := testutil.NewMockEntity(def.EntityTypePlayer)
	enemy := testutil.NewMockEntity(def.EntityTypeEnemy)

	store.Add(background)
	store.Add(enemy)
	store.Add(player)

	// Collect iteration order
	var order []def.EntityType
	for _, entity := range store.IterateForUpdate() {
		order = append(order, entity.Type())
	}

	// Verify order matches EntityTypes priority
	// (Environment, Player, Team, Enemy, Obstacle, Background)
	want := []def.EntityType{
		def.EntityTypePlayer,
		def.EntityTypeEnemy,
		def.EntityTypeBackground,
	}

	if len(order) != len(want) {
		t.Fatalf("iteration count = %d, want %d", len(order), len(want))
	}

	for i, gotType := range order {
		if gotType != want[i] {
			t.Errorf("order[%d] = %v, want %v", i, gotType, want[i])
		}
	}
}

func TestEntityStoreIterateForDrawOrder(t *testing.T) {
	store := play.NewEntityStore()

	// Add entities
	background := testutil.NewMockEntity(def.EntityTypeBackground)
	player := testutil.NewMockEntity(def.EntityTypePlayer)
	enemy := testutil.NewMockEntity(def.EntityTypeEnemy)

	store.Add(background)
	store.Add(enemy)
	store.Add(player)

	// Collect iteration order
	var order []def.EntityType
	for _, entity := range store.IterateForDraw() {
		order = append(order, entity.Type())
	}

	// Verify reverse order (Background first, Player last)
	want := []def.EntityType{
		def.EntityTypeBackground,
		def.EntityTypeEnemy,
		def.EntityTypePlayer,
	}

	if len(order) != len(want) {
		t.Fatalf("iteration count = %d, want %d", len(order), len(want))
	}

	for i, gotType := range order {
		if gotType != want[i] {
			t.Errorf("order[%d] = %v, want %v", i, gotType, want[i])
		}
	}
}

func TestEntityStoreRemovesMarkedEntities(t *testing.T) {
	store := play.NewEntityStore()

	entity1 := testutil.NewMockEntity(def.EntityTypeEnemy)
	entity2 := testutil.NewMockEntity(def.EntityTypeEnemy)
	entity2.Removed = true // Mark for removal
	entity3 := testutil.NewMockEntity(def.EntityTypeEnemy)

	store.Add(entity1)
	store.Add(entity2)
	store.Add(entity3)

	// First iteration should remove entity2
	count := 0
	for range store.IterateForUpdate() {
		count++
	}

	got := count
	want := 2 // entity1 and entity3, entity2 was removed

	if got != want {
		t.Errorf("after removal, entity count = %d, want %d", got, want)
	}

	// Second iteration should still have 2 entities
	count = 0
	for range store.IterateForUpdate() {
		count++
	}

	if count != want {
		t.Errorf("second iteration count = %d, want %d (removal should be permanent)", count, want)
	}
}

func TestEntityStoreSameTypeOrderPreserved(t *testing.T) {
	store := play.NewEntityStore()

	// Add multiple enemies in specific order
	enemies := []*testutil.MockEntity{
		testutil.NewMockEntity(def.EntityTypeEnemy),
		testutil.NewMockEntity(def.EntityTypeEnemy),
		testutil.NewMockEntity(def.EntityTypeEnemy),
	}

	for _, e := range enemies {
		store.Add(e)
	}

	// Collect iteration order
	var collected []*testutil.MockEntity
	for _, entity := range store.IterateForUpdate() {
		collected = append(collected, entity.(*testutil.MockEntity))
	}

	// Verify FIFO order within the type
	if len(collected) != len(enemies) {
		t.Fatalf("collected count = %d, want %d", len(collected), len(enemies))
	}

	for i, gotEntity := range collected {
		if gotEntity != enemies[i] {
			t.Errorf("order[%d]: got different entity, want same instance (FIFO violated)", i)
		}
	}
}

func TestEntityStoreTypeGet(t *testing.T) {
	store := play.NewEntityStore()

	// Add entities of different types
	player := testutil.NewMockEntity(def.EntityTypePlayer)
	enemy := testutil.NewMockEntity(def.EntityTypeEnemy)
	background := testutil.NewMockEntity(def.EntityTypeBackground)

	store.Add(player)
	store.Add(enemy)
	store.Add(background)

	// Get entities by type
	gotPlayers := store.Get(def.EntityTypePlayer)
	gotEnemies := store.Get(def.EntityTypeEnemy)
	gotBackgrounds := store.Get(def.EntityTypeBackground)

	if len(gotPlayers) != 1 || gotPlayers[0] != player {
		t.Errorf("Get(EntityTypePlayer) = %v, want [%v]", gotPlayers, player)
	}

	if len(gotEnemies) != 1 || gotEnemies[0] != enemy {
		t.Errorf("Get(EntityTypeEnemy) = %v, want [%v]", gotEnemies, enemy)
	}

	if len(gotBackgrounds) != 1 || gotBackgrounds[0] != background {
		t.Errorf("Get(EntityTypeBackground) = %v, want [%v]", gotBackgrounds, background)
	}
}

// TestGetDuringActSeesObstacles reproduces the Chaser segfault: an Enemy
// entity calls scene.Entities().Get(Obstacle) inside its Act, while several
// Obstacle entities are present. With the old goroutine-based IterateForUpdate
// the goroutine raced against the Get call and could corrupt the returned
// slice. The fix (synchronous collection before Act) makes this deterministic.
func TestGetDuringActSeesObstacles(t *testing.T) {
	store := play.NewEntityStore()

	obs1 := testutil.NewMockEntity(def.EntityTypeObstacle)
	obs2 := testutil.NewMockEntity(def.EntityTypeObstacle)
	obs3 := testutil.NewMockEntity(def.EntityTypeObstacle)
	store.Add(obs1)
	store.Add(obs2)
	store.Add(obs3)

	spy := newSpyEntity(def.EntityTypeEnemy, def.EntityTypeObstacle)
	store.Add(spy)

	// Use MockScene so Act receives a valid def.Scene with the obstacles.
	mockScene := testutil.NewMockScene()
	mockScene.Entities().Add(obs1)
	mockScene.Entities().Add(obs2)
	mockScene.Entities().Add(obs3)

	// Manually drive Act the same way scene.Update does after IterateForUpdate.
	for _, e := range store.IterateForUpdate() {
		e.Act(mockScene)
	}

	// The spy should have seen all three obstacles and none should be nil.
	if len(spy.seen) != 3 {
		t.Fatalf("spy saw %d obstacle(s), want 3", len(spy.seen))
	}
	for i, e := range spy.seen {
		if e == nil {
			t.Errorf("spy.seen[%d] is nil — Get returned a nil entity during Act", i)
		}
	}
}
