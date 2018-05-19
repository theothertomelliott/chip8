package main

import (
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/theothertomelliott/chip8/chip8"
	"golang.org/x/image/colornames"
)

const sizeX, sizeY = 64, 32
const screenWidth, screenHeight = float64(1024), float64(768)

var win *pixelgl.Window

func main() {
	pixelgl.Run(run)
}

func run() {

	// Set up render system and register input callbacks
	setupGraphics()

	myChip8 := &chip8.Chip8{}

	// Initialize the Chip8 system and load the game into the memory
	myChip8.Initialize()
	err := myChip8.LoadGame("data/pong.ch8")
	if err != nil {
		log.Fatal(err)
	}
	// Emulation loop
	for !win.Closed() {
		// Emulate one cycle
		if err := myChip8.EmulateCycle(); err != nil {
			log.Fatal(err)
		}

		// If the draw flag is set, update the screen
		if myChip8.DrawFlag() {
			drawGraphics(myChip8.GetGraphics())
		}

		handleKeys(myChip8)
	}
}

func setupGraphics() {
	cfg := pixelgl.WindowConfig{
		Title:  "Chip8",
		Bounds: pixel.R(0, 0, screenWidth, screenHeight),
		VSync:  true,
	}
	var err error
	win, err = pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}
}

func handleKeys(myChip8 *chip8.Chip8) {
	// Store key press state (Press and Release)
	var keyByIndex = []pixelgl.Button{
		pixelgl.Key1, pixelgl.Key2, pixelgl.Key3, pixelgl.Key4,
		pixelgl.KeyQ, pixelgl.KeyW, pixelgl.KeyE, pixelgl.KeyR,
		pixelgl.KeyA, pixelgl.KeyS, pixelgl.KeyD, pixelgl.KeyF,
		pixelgl.KeyZ, pixelgl.KeyX, pixelgl.KeyC, pixelgl.KeyV,
	}

	for index, key := range keyByIndex {
		if win.JustPressed(key) {
			myChip8.SetKey(byte(index), true)
		}
		if win.JustReleased(key) {
			myChip8.SetKey(byte(index), false)
		}
	}
}

func drawGraphics(graphics [64 * 32]byte) {
	win.Clear(colornames.Black)
	imd := imdraw.New(nil)
	imd.Color = pixel.RGB(1, 1, 1)
	screenWidth := win.Bounds().W()
	width, height := screenWidth/sizeX, screenHeight/sizeY
	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			if graphics[(31-y)*64+x] == 1 {
				imd.Push(pixel.V(width*float64(x), height*float64(y)))
				imd.Push(pixel.V(width*float64(x)+width, height*float64(y)+height))
				imd.Rectangle(0)
			}
		}
	}
	imd.Draw(win)
	win.Update()
}
