package play

// GameMode represents the current game state/screen
type GameMode int

const (
	GameModeIntro GameMode = iota
	GameModePlay
	GameModeGameOver
	GameModeTransition
	GameModeComplete
)
