package def

// GameStateReader provides read-only access to game-wide state for UI entities.
type GameStateReader interface {
	GetWave() int
	GetScore() int
}

// Scorer is implemented by entities that award points when killed.
type Scorer interface {
	Entity
	ScoreValue() int
}
