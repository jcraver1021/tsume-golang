package testutil

import (
	"sync"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/def"
)

// --- MockScene ---

// MockScene is a test implementation of def.Scene.
// Safe for concurrent use; all entity access is guarded by an RWMutex.
type MockScene struct {
	mu       sync.RWMutex
	entities []def.Entity
	width    int
	height   int
	tick     int
}

func NewMockScene() *MockScene {
	return &MockScene{
		width:  def.ScreenWidth,
		height: def.ScreenHeight,
	}
}

func NewMockSceneWithSize(width, height int) *MockScene {
	return &MockScene{width: width, height: height}
}

func (m *MockScene) Width() int  { return m.width }
func (m *MockScene) Height() int { return m.height }
func (m *MockScene) Tick() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.tick
}

func (m *MockScene) IncrementTick() {
	m.mu.Lock()
	m.tick++
	m.mu.Unlock()
}

func (m *MockScene) Entities() def.EntityCollection {
	return &MockEntityCollection{scene: m}
}

// GetEntities returns a snapshot of the entity slice for test assertions.
func (m *MockScene) GetEntities() []def.Entity {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]def.Entity, len(m.entities))
	copy(out, m.entities)
	return out
}

func (m *MockScene) EntityCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entities)
}

func (m *MockScene) Clear() {
	m.mu.Lock()
	m.entities = nil
	m.mu.Unlock()
}

// --- MockEntityCollection ---

// MockEntityCollection is a test implementation of def.EntityCollection.
// NOTE: Get is O(n) — it scans all entities to filter by type. Fine for
// small test fixtures; don't use MockScene with hundreds of entities.
type MockEntityCollection struct {
	scene *MockScene
}

func (m *MockEntityCollection) Add(e def.Entity) {
	m.scene.mu.Lock()
	m.scene.entities = append(m.scene.entities, e)
	m.scene.mu.Unlock()
}

func (m *MockEntityCollection) Get(entityType def.EntityType) []def.Entity {
	m.scene.mu.RLock()
	defer m.scene.mu.RUnlock()
	var result []def.Entity
	for _, e := range m.scene.entities {
		if e.Type() == entityType {
			result = append(result, e)
		}
	}
	return result
}

func (m *MockEntityCollection) IterateForUpdate() []def.Entity {
	return m.scene.GetEntities()
}

func (m *MockEntityCollection) IterateForDraw() []def.Entity {
	snapshot := m.scene.GetEntities()
	reversed := make([]def.Entity, len(snapshot))
	for i, e := range snapshot {
		reversed[len(snapshot)-1-i] = e
	}
	return reversed
}

// --- MockEntity ---

// MockEntity is a minimal implementation of def.Entity for use in tests.
type MockEntity struct {
	EntityType          def.EntityType
	X, Y, Width, Height int
	Removed             bool
}

func NewMockEntity(entityType def.EntityType) *MockEntity {
	return &MockEntity{EntityType: entityType, Width: 10, Height: 10}
}

func (m *MockEntity) Type() def.EntityType   { return m.EntityType }
func (m *MockEntity) Location() (int, int)   { return m.X, m.Y }
func (m *MockEntity) Dimensions() (int, int) { return m.Width, m.Height }
func (m *MockEntity) Act(_ def.Scene)        {}
func (m *MockEntity) Draw(_ *ebit.Image)     {}
func (m *MockEntity) CanBeRemoved() bool     { return m.Removed }
func (m *MockEntity) BoundingBoxOverlaps(other def.Entity) bool {
	ox, oy := other.Location()
	ow, oh := other.Dimensions()
	return !(m.X+m.Width <= ox || m.X >= ox+ow || m.Y+m.Height <= oy || m.Y >= oy+oh)
}

// --- MockMortalEntity ---

// MockMortalEntity extends MockEntity with def.Mortal. IsDead becomes true
// after MarkAsDead is called; entities added in MarkAsDead are recorded in
// the provided scene.
type MockMortalEntity struct {
	*MockEntity
	dead        bool
	DeathEffect def.DeathEffect
}

func NewMockMortalEntity(entityType def.EntityType) *MockMortalEntity {
	return &MockMortalEntity{MockEntity: NewMockEntity(entityType)}
}

func (m *MockMortalEntity) IsDead() bool                    { return m.dead }
func (m *MockMortalEntity) GetDeathEffect() def.DeathEffect { return m.DeathEffect }
func (m *MockMortalEntity) MarkAsDead(_ def.Scene)          { m.dead = true }

// --- MockDamageableEntity ---

// MockDamageableEntity extends MockEntity with def.Damageable. HP floors at 0.
type MockDamageableEntity struct {
	*MockEntity
	hp    int
	maxHP int
}

func NewMockDamageableEntity(entityType def.EntityType, maxHP int) *MockDamageableEntity {
	return &MockDamageableEntity{MockEntity: NewMockEntity(entityType), hp: maxHP, maxHP: maxHP}
}

func (m *MockDamageableEntity) CurrentHP() int { return m.hp }
func (m *MockDamageableEntity) MaxHP() int     { return m.maxHP }
func (m *MockDamageableEntity) TakeDamage(amount int) {
	m.hp -= amount
	if m.hp < 0 {
		m.hp = 0
	}
}

// --- MockImpulsableEntity ---

// MockImpulsableEntity extends MockEntity with def.Impulsable. It records
// the most recent impulse applied for assertion in tests.
type MockImpulsableEntity struct {
	*MockEntity
	LastDVX, LastDVY float64
	ImpulseCount     int
}

func NewMockImpulsableEntity(entityType def.EntityType) *MockImpulsableEntity {
	return &MockImpulsableEntity{MockEntity: NewMockEntity(entityType)}
}

func (m *MockImpulsableEntity) ApplyImpulse(dvx, dvy float64) {
	m.LastDVX = dvx
	m.LastDVY = dvy
	m.ImpulseCount++
}

// --- MockGameStateReader ---

// MockGameStateReader is a test implementation of def.GameStateReader with
// settable Wave and Score fields.
type MockGameStateReader struct {
	Wave  int
	Score int
}

func (m *MockGameStateReader) GetWave() int  { return m.Wave }
func (m *MockGameStateReader) GetScore() int { return m.Score }

// --- MockAmmoPlayer ---

// MockAmmoPlayer is a player entity that satisfies the HUD's internal ammoReader
// interface (SecondaryAmmo() (current, max int, hasWeapon bool)).
type MockAmmoPlayer struct {
	*MockEntity
	CurrentAmmo int
	MaxAmmo     int
	HasWeapon   bool
}

func NewMockAmmoPlayer() *MockAmmoPlayer {
	return &MockAmmoPlayer{MockEntity: NewMockEntity(def.EntityTypePlayer)}
}

func (m *MockAmmoPlayer) SecondaryAmmo() (int, int, bool) {
	return m.CurrentAmmo, m.MaxAmmo, m.HasWeapon
}
