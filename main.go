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
	setupInput()

	myChip8 := &chip8.Chip8{}

	// Initialize the Chip8 system and load the game into the memory
	myChip8.Initialize()
	err := myChip8.LoadGame("data/pong.rom")
	if err != nil {
		panic(err)
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

		// Store key press state (Press and Release)
		myChip8.SetKeys()

		// TODO: Handle interrupts
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

func setupInput() {

}

func drawGraphics(graphics [64 * 32]byte) {
	win.Clear(colornames.Black)
	imd := imdraw.New(nil)
	imd.Color = pixel.RGB(1, 0, 0)
	screenWidth := win.Bounds().W()
	width, height := screenWidth/sizeX, screenWidth/sizeY
	for x := 0; x < 64; x++ {
		for y := 0; y < 32; y++ {
			if graphics[x*32+y] == 1 {
				imd.Push(pixel.V(width*float64(x), height*float64(y)))
				imd.Push(pixel.V(width*float64(x)+width, height*float64(y)+height))
				imd.Rectangle(0)
			}
		}
	}
	imd.Draw(win)
	win.Update()
}
