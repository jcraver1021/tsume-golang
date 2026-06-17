package play

import (
	"slices"

	"tsumegolang/game/starshot/def"
	ds "tsumegolang/pkg/ds/basic"
)

var entityTypeCapacity = map[def.EntityType]int{
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

// IterateForUpdate iterates entities in their natural order of types for game logic updates
// (player first, background last).
func (s *EntityStore) IterateForUpdate() <-chan def.Entity {
	ch := make(chan def.Entity)
	go func() {
		for _, entityType := range def.EntityTypes {
			deque := s.entityMap[entityType]

			for range deque.Len() {
				e, ok := deque.PopFront()
				if ok && !e.CanBeRemoved() {
					ch <- e
					deque.PushBack(e)
				}
			}
		}

		close(ch)
	}()

	return ch
}

// IterateForDraw iterates entities in reverse order of their types to ensure correct rendering order
// (background first, player last).
func (s *EntityStore) IterateForDraw() <-chan def.Entity {
	ch := make(chan def.Entity)

	go func() {
		reversedTypes := make([]def.EntityType, len(def.EntityTypes))
		copy(reversedTypes, def.EntityTypes)
		slices.Reverse(reversedTypes)

		for _, entityType := range reversedTypes {
			deque := s.entityMap[entityType]

			for range deque.Len() {
				e, ok := deque.PopBack()
				if ok && !e.CanBeRemoved() {
					ch <- e
					deque.PushFront(e)
				}
			}
		}

		close(ch)
	}()

	return ch
}
