package def

// Screen dimensions shared across all entities and scenes.
const (
	ScreenWidth  = 480
	ScreenHeight = 640
)

// OnScreen indicates an object's position relative to the screen boundary.
type OnScreen int

const (
	Fully     OnScreen = iota // entirely within screen bounds
	Partially                 // overlapping the edge
	OffScreen                 // entirely outside
)

// Scene is the context passed into every entity's Act call.
// It exposes the playfield dimensions, the entity collection for spawning
// new entities, and a monotonic tick counter for animation timing.
type Scene interface {
	Width() int
	Height() int
	Entities() EntityCollection
	Tick() int
}
