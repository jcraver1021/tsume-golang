package def

// EntityCollection provides access to entities without exposing implementation details
type EntityCollection interface {
	Add(Entity)
	Get(EntityType) []Entity
	IterateForUpdate() <-chan Entity
	IterateForDraw() <-chan Entity
}
