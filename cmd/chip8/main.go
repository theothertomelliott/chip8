package main

import (
	"log"
	"os"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/theothertomelliott/chip8"
	"golang.org/x/image/colornames"
)

const (
	cyclesPerSecond           = 300
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

	// Open the ROM specified as argument ready to load
	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Create a CHIP-8 machine and load the ROM file
	myChip8, err := chip8.New(file)
	_ = file.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Should trace logging be output?
	var trace bool

	// Emulation loop
	for !win.Closed() {
		if win.Pressed(pixelgl.KeyEscape) {
			return
		}
		// Toggle operation tracing
		if win.JustPressed(pixelgl.KeyT) {
			trace = !trace
		}

		// Emulate one cycle
		result, err := myChip8.EmulateCycle()
		if err != nil {
			log.Fatalf("0x%X> %v", result.Before.PC, err)
		}
		if trace {
			log.Printf("0x%X> (0x%X) %s", result.Before.PC, result.Opcode, result.Pseudo)
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
	keyByIndex = map[uint16]pixelgl.Button{
		0x1: pixelgl.Key1, 0x2: pixelgl.Key2, 0x3: pixelgl.Key3, 0xC: pixelgl.Key4,
		0x4: pixelgl.KeyQ, 0x5: pixelgl.KeyW, 0x6: pixelgl.KeyE, 0xD: pixelgl.KeyR,
		0x7: pixelgl.KeyA, 0x8: pixelgl.KeyS, 0x9: pixelgl.KeyD, 0xE: pixelgl.KeyF,
		0xA: pixelgl.KeyZ, 0x0: pixelgl.KeyX, 0xB: pixelgl.KeyC, 0xF: pixelgl.KeyV,
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
			myChip8.SetKeyDown(byte(index))
		}

		if keysDown[index] == nil {
			continue
		}
		select {
		case <-keysDown[index].C:
			myChip8.SetKeyDown(byte(index))
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
