package testutil

import (
	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
)

// MockScene is a test implementation of def.Scene
type MockScene struct {
	entities []def.Entity
	width    int
	height   int
	tick     int
}

// NewMockScene creates a new mock scene with default screen dimensions
func NewMockScene() *MockScene {
	return &MockScene{
		entities: []def.Entity{},
		width:    def.ScreenWidth,
		height:   def.ScreenHeight,
		tick:     0,
	}
}

// NewMockSceneWithSize creates a mock scene with custom dimensions
func NewMockSceneWithSize(width, height int) *MockScene {
	return &MockScene{
		entities: []def.Entity{},
		width:    width,
		height:   height,
		tick:     0,
	}
}

// Width returns the scene width
func (m *MockScene) Width() int { return m.width }

// Height returns the scene height
func (m *MockScene) Height() int { return m.height }

// Tick returns the current tick counter
func (m *MockScene) Tick() int { return m.tick }

// IncrementTick advances the tick counter (for testing)
func (m *MockScene) IncrementTick() { m.tick++ }

// Entities returns the mock entity collection
func (m *MockScene) Entities() def.EntityCollection {
	return &MockEntityCollection{scene: m}
}

// GetEntities returns the underlying entity slice for test assertions
func (m *MockScene) GetEntities() []def.Entity {
	return m.entities
}

// EntityCount returns the number of entities in the scene
func (m *MockScene) EntityCount() int {
	return len(m.entities)
}

// Clear removes all entities from the scene
func (m *MockScene) Clear() {
	m.entities = []def.Entity{}
}

// MockEntityCollection is a test implementation of def.EntityCollection
type MockEntityCollection struct {
	scene *MockScene
}

// Add adds an entity to the collection
func (m *MockEntityCollection) Add(e def.Entity) {
	m.scene.entities = append(m.scene.entities, e)
}

// Get returns entities of the specified type
func (m *MockEntityCollection) Get(entityType def.EntityType) []def.Entity {
	var result []def.Entity
	for _, e := range m.scene.entities {
		if e.Type() == entityType {
			result = append(result, e)
		}
	}
	return result
}

// IterateForUpdate iterates entities in forward order
func (m *MockEntityCollection) IterateForUpdate() <-chan def.Entity {
	ch := make(chan def.Entity)
	go func() {
		for _, e := range m.scene.entities {
			ch <- e
		}
		close(ch)
	}()
	return ch
}

// IterateForDraw iterates entities in reverse order
func (m *MockEntityCollection) IterateForDraw() <-chan def.Entity {
	ch := make(chan def.Entity)
	go func() {
		for i := len(m.scene.entities) - 1; i >= 0; i-- {
			ch <- m.scene.entities[i]
		}
		close(ch)
	}()
	return ch
}

// MockEntity is a simple test implementation of def.Entity
type MockEntity struct {
	EntityType          def.EntityType
	X, Y, Width, Height int
	Removed             bool
}

// NewMockEntity creates a basic mock entity
func NewMockEntity(entityType def.EntityType) *MockEntity {
	return &MockEntity{
		EntityType: entityType,
		X:          0,
		Y:          0,
		Width:      10,
		Height:     10,
		Removed:    false,
	}
}

// Type returns the entity type
func (m *MockEntity) Type() def.EntityType { return m.EntityType }

// Location returns the entity position
func (m *MockEntity) Location() (x, y int) { return m.X, m.Y }

// Dimensions returns the entity size
func (m *MockEntity) Dimensions() (width, height int) { return m.Width, m.Height }

// BoundingBoxOverlaps implements basic AABB collision
func (m *MockEntity) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(m.X+m.Width < ox || m.X > ox+ow || m.Y+m.Height < oy || m.Y > oy+oh)
}

// Act does nothing in the mock
func (m *MockEntity) Act(scene def.Scene) {}

// Draw does nothing in the mock
func (m *MockEntity) Draw(img *ebit.Image) {}

// CanBeRemoved returns the Removed flag
func (m *MockEntity) CanBeRemoved() bool { return m.Removed }
