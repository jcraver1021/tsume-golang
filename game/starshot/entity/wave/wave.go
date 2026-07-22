package wave

import (
	"tsumegolang/game/starshot/def"
)

func LoadWave(b def.Scene, waveNumber int) {
	switch waveNumber {
	case 1:
		b.Entities().Add(NewWave1())
	}
}
