package play

import (
	"slices"

	"tsumegolang/game/starshot/def"
	ds "tsumegolang/pkg/ds/basic"
)

var entityTypeCapacity = map[def.EntityType]int{
	def.EntityTypeUI:          8,
	def.EntityTypeEnvironment: 8,
	def.EntityTypePlayer:      1,
	def.EntityTypeTeam:        8,
	def.EntityTypeEnemy:       64,
	def.EntityTypeObstacle:    64,
	def.EntityTypeBackground:  64,
}

type EntityStore struct {
	entityMap map[def.EntityType]*ds.Deque[def.Entity]
}

func NewEntityStore() *EntityStore {
	store := &EntityStore{
		entityMap: make(map[def.EntityType]*ds.Deque[def.Entity]),
	}
	for _, entityType := range def.EntityTypes {
		store.entityMap[entityType] = ds.NewDeque(ds.WithDequeCapacity[def.Entity](entityTypeCapacity[entityType]))
	}
	return store
}

func (s *EntityStore) Add(e def.Entity) {
	s.entityMap[e.Type()].PushBack(e)
}

func (s *EntityStore) Get(entityType def.EntityType) []def.Entity {
	return s.entityMap[entityType].ToSlice()
}

// IterateForUpdate returns entities to update, pruning removable ones first.
// Removal is a separate pass before any Act calls, so Act can safely call
// scene.Entities().Get without racing against deque mutation.
func (s *EntityStore) IterateForUpdate() []def.Entity {
	result := make([]def.Entity, 0, 64)
	for _, entityType := range def.EntityTypes {
		deque := s.entityMap[entityType]
		for range deque.Len() {
			e, ok := deque.PopFront()
			if ok && !e.CanBeRemoved() {
				deque.PushBack(e)
				result = append(result, e)
			}
		}
	}
	return result
}

// IterateForDraw returns entities to draw in back-to-front order (background
// drawn first so it renders behind everything).
func (s *EntityStore) IterateForDraw() []def.Entity {
	reversedTypes := make([]def.EntityType, len(def.EntityTypes))
	copy(reversedTypes, def.EntityTypes)
	slices.Reverse(reversedTypes)

	result := make([]def.Entity, 0, 64)
	for _, entityType := range reversedTypes {
		deque := s.entityMap[entityType]
		for range deque.Len() {
			e, ok := deque.PopBack()
			if ok && !e.CanBeRemoved() {
				deque.PushFront(e)
				result = append(result, e)
			}
		}
	}
	return result
}
