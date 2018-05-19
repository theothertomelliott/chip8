package main

import (
	"log"

	"github.com/theothertomelliott/chip8/chip8"
)

func main() {
	// Set up render system and register input callbacks
	setupGraphics()
	setupInput()

	myChip8 := chip8.Chip8{}

	// Initialize the Chip8 system and load the game into the memory
	myChip8.Initialize()
	err := myChip8.LoadGame("data/pong.rom")
	if err != nil {
		panic(err)
	}
	// Emulation loop
	for true {
		// Emulate one cycle
		if err := myChip8.EmulateCycle(); err != nil {
			log.Fatal(err)
		}

		// If the draw flag is set, update the screen
		if myChip8.DrawFlag() {
			drawGraphics()
		}

		// Store key press state (Press and Release)
		myChip8.SetKeys()

		// TODO: Handle interrupts
	}
}

func setupGraphics() {

}

func setupInput() {

}

func drawGraphics() {

}
