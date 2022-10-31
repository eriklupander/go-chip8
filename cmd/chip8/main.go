package main

import (
	"github.com/eriklupander/go-chip8/internal/app"
	"github.com/eriklupander/go-chip8/internal/app/runtime"
	"github.com/hajimehoshi/ebiten/v2"
	"log"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixMilli())
}

func main() {

	uiRuntime := runtime.NewUIRuntime()
	emul := app.NewEmulator()

	go emul.Run(uiRuntime, "roms/pong.ch8")

	if err := ebiten.RunGame(uiRuntime); err != nil {
		log.Fatal(err)
	}
}
