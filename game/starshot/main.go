package main

import (
	"log"

	ebit "github.com/hajimehoshi/ebiten/v2"
	"tsumegolang/game/starshot/play"
)

func main() {
	game := play.NewGame()
	if err := ebit.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
