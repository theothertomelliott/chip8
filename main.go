package main

import (
	"log"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/theothertomelliott/chip8/chip8"
	"golang.org/x/image/colornames"
)

const (
	cyclesPerSecond           = 1000
	sizeX, sizeY              = 64, 32
	screenWidth, screenHeight = float64(1024), float64(768)
	keyRepeatDuration         = time.Second / 5
)

var win *pixelgl.Window

func main() {
	pixelgl.Run(run)
}

func run() {

	ticker := time.NewTicker(time.Second / cyclesPerSecond)
	defer ticker.Stop()

	// Set up render system and register input callbacks
	setupGraphics()

	myChip8 := &chip8.Chip8{}

	// Initialize the Chip8 system and load the game into the memory
	myChip8.Initialize()
	err := myChip8.LoadGame(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	// Emulation loop
	for !win.Closed() {
		if win.Pressed(pixelgl.KeyEscape) {
			return
		}

		// Emulate one cycle
		if err := myChip8.EmulateCycle(); err != nil {
			log.Fatal(err)
		}

		// If the draw flag is set, update the screen
		if myChip8.DrawFlag() {
			drawGraphics(myChip8.GetGraphics())
		} else {
			win.UpdateInput()
		}

		handleKeys(myChip8)

		// Wait for the next tick
		<-ticker.C
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

// Store key press state (Press and Release)
var (
	keyByIndex = []pixelgl.Button{
		pixelgl.Key1, pixelgl.Key2, pixelgl.Key3, pixelgl.Key4,
		pixelgl.KeyQ, pixelgl.KeyW, pixelgl.KeyE, pixelgl.KeyR,
		pixelgl.KeyA, pixelgl.KeyS, pixelgl.KeyD, pixelgl.KeyF,
		pixelgl.KeyZ, pixelgl.KeyX, pixelgl.KeyC, pixelgl.KeyV,
	}
	keysDown [16]*time.Ticker
)

func handleKeys(myChip8 *chip8.Chip8) {

	for index, key := range keyByIndex {
		if win.JustReleased(key) {
			if keysDown[index] != nil {
				keysDown[index].Stop()
				keysDown[index] = nil
			}
		} else if win.JustPressed(key) {
			if keysDown[index] == nil {
				keysDown[index] = time.NewTicker(keyRepeatDuration)
			}
			myChip8.SetKey(byte(index), true)
		}

		if keysDown[index] == nil {
			continue
		}
		select {
		case <-keysDown[index].C:
			myChip8.SetKey(byte(index), true)
		default:
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
