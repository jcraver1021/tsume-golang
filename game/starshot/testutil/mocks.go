package testutil

import (
	"tsumegolang/game/starshot/def"
)

// MockScene is a test implementation of def.Scene
type MockScene struct {
	entities []def.Entity
	width    int
	height   int
}

// NewMockScene creates a new mock scene with default screen dimensions
func NewMockScene() *MockScene {
	return &MockScene{
		entities: []def.Entity{},
		width:    def.ScreenWidth,
		height:   def.ScreenHeight,
	}
}

// NewMockSceneWithSize creates a mock scene with custom dimensions
func NewMockSceneWithSize(width, height int) *MockScene {
	return &MockScene{
		entities: []def.Entity{},
		width:    width,
		height:   height,
	}
}

// Width returns the scene width
func (m *MockScene) Width() int { return m.width }

// Height returns the scene height
func (m *MockScene) Height() int { return m.height }

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
