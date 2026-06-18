package play_test

import (
	"testing"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
	"tsumegolang/game/starshot/play"
)

// mockEntity is a minimal test entity
type mockEntity struct {
	entityType def.EntityType
	removed    bool
}

func newMockEntity(t def.EntityType) *mockEntity {
	return &mockEntity{entityType: t}
}

func (m *mockEntity) Type() def.EntityType                      { return m.entityType }
func (m *mockEntity) Location() (x, y int)                      { return 0, 0 }
func (m *mockEntity) Dimensions() (width, height int)           { return 1, 1 }
func (m *mockEntity) BoundingBoxOverlaps(other def.Entity) bool { return false }
func (m *mockEntity) Act(s def.Scene)                           {}
func (m *mockEntity) Draw(img *ebit.Image)                      {}
func (m *mockEntity) CanBeRemoved() bool                        { return m.removed }

func TestEntityStoreImplementsInterface(t *testing.T) {
	store := play.NewEntityStore()
	var _ def.EntityCollection = store
}

func TestEntityStoreAdd(t *testing.T) {
	store := play.NewEntityStore()
	entity := newMockEntity(def.EntityTypePlayer)

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
	background := newMockEntity(def.EntityTypeBackground)
	player := newMockEntity(def.EntityTypePlayer)
	enemy := newMockEntity(def.EntityTypeEnemy)

	store.Add(background)
	store.Add(enemy)
	store.Add(player)

	// Collect iteration order
	var order []def.EntityType
	for entity := range store.IterateForUpdate() {
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
	background := newMockEntity(def.EntityTypeBackground)
	player := newMockEntity(def.EntityTypePlayer)
	enemy := newMockEntity(def.EntityTypeEnemy)

	store.Add(background)
	store.Add(enemy)
	store.Add(player)

	// Collect iteration order
	var order []def.EntityType
	for entity := range store.IterateForDraw() {
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

	entity1 := newMockEntity(def.EntityTypeEnemy)
	entity2 := newMockEntity(def.EntityTypeEnemy)
	entity2.removed = true // Mark for removal
	entity3 := newMockEntity(def.EntityTypeEnemy)

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
	enemies := []*mockEntity{
		newMockEntity(def.EntityTypeEnemy),
		newMockEntity(def.EntityTypeEnemy),
		newMockEntity(def.EntityTypeEnemy),
	}

	for _, e := range enemies {
		store.Add(e)
	}

	// Collect iteration order
	var collected []*mockEntity
	for entity := range store.IterateForUpdate() {
		collected = append(collected, entity.(*mockEntity))
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
	player := newMockEntity(def.EntityTypePlayer)
	enemy := newMockEntity(def.EntityTypeEnemy)
	background := newMockEntity(def.EntityTypeBackground)

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
