package chip8

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"
)

// Chip8 emulates a CHIP-8 machine.
// An initialized instance can be created with New()
type Chip8 struct {
	opcode uint16

	memory [4096]byte

	// The Chip 8 has 15 8-bit general purpose registers named V0,V1 up to VE.
	// The 16th register is used  for the ‘carry flag’.
	// Eight bits is one byte so we can use an unsigned char for this purpose
	V [16]byte

	// There is an Index register I and a program counter (pc) which can have a value from 0x000 to 0xFFF
	I  uint16
	pc uint16

	// The systems memory map:
	// 0x000-0x1FF - Chip 8 interpreter (contains font set in emu)
	// 0x050-0x0A0 - Used for the built in 4x5 pixel font set (0-F)
	// 0x200-0xFFF - Program ROM and work RAM

	// The graphics system: The chip 8 has one instruction that draws sprite to the screen.
	// Drawing is done in XOR mode and if a pixel is turned off as a result of drawing, the VF register is set.
	// This is used for collision detection.
	// The graphics of the Chip 8 are black and white and the screen has a total of 2048 pixels (64 x 32).
	// This can easily be implemented using an array that hold the pixel state (1 or 0):
	gfx [64 * 32]byte

	// Interupts and hardware registers.
	// The Chip 8 has none, but there are two timer registers that count at 60 Hz.
	// When set above zero they will count down to zero.
	delayTimer byte
	soundTimer byte

	// It is important to know that the Chip 8 instruction set has opcodes that allow the program to jump to a certain address or call a subroutine.
	// While the specification don’t mention a stack, you will need to implement one as part of the interpreter yourself.
	// The stack is used to remember the current location before a jump is performed.
	// So anytime you perform a jump or call a subroutine, store the program counter in the stack before proceeding.
	// The system has 16 levels of stack and in order to remember which level of the stack is used, you need to implement a stack pointer (sp).
	stack [16]uint16
	sp    uint16

	// Finally, the Chip 8 has a HEX based keypad (0x0-0xF), you can use an array to store the current state of the key.
	key [16]byte

	// True iff the screen must be drawn
	drawFlag bool

	timerClock *time.Ticker

	opcodes map[uint16]opcodeHandler

	beepOut chan struct{}
}

// Result records the actions performed when handling an opcode.
// This can be used for logging or other telemetry.
type Result struct {
	Opcode     uint16
	OpcodeType string
	Pseudo     string

	Before ResultState
	After  ResultState
}

// ResultState provides a snapshot of CPU state
// for use in tracing.
type ResultState struct {
	PC uint16
	V  [16]byte
}

// New creates a new CHIP-8 machine in a starting condition.
// Empty registers, stack and display, zeroed timers and
// memory populated with font data and the contents of a ROM
// provided in an io.Reader.
//
// The Chip8 instance returned will be ready to start processing
// opcodes with calls to ExecuteCycle.
func New(rom io.Reader) (*Chip8, error) {
	c := &Chip8{}
	c.initialize()

	err := c.loadROM(rom)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Chip8) initialize() {
	// Set up opcode mapping
	c.registerOpcodeHandlers()

	// Initialize registers and memory once
	c.pc = 0x200 // Program counter starts at 0x200
	c.opcode = 0 // Reset current opcode
	c.I = 0      // Reset index register
	c.sp = 0     // Reset stack pointer

	// Clear display
	c.gfx = [64 * 32]byte{}
	// Clear stack
	c.stack = [16]uint16{}
	// Clear registers V0-VF
	c.V = [16]byte{}
	// Clear memory
	c.memory = [4096]byte{}

	// Load fontset
	for i := 0; i < len(chip8Fontset); i++ {
		c.memory[i] = chip8Fontset[i]
	}
	// Reset timers
	c.delayTimer = 0
	c.soundTimer = 0

	// Set up output for beeps
	c.beepOut = make(chan struct{})

	// Create a ticker at 60Hz
	c.timerClock = time.NewTicker(time.Second / 60)
}

// loadROM loads a ROM into memory from an io.Reader
func (c *Chip8) loadROM(rom io.Reader) error {
	bytes, err := ioutil.ReadAll(rom)
	if err != nil {
		return err
	}

	for i := 0; i < len(bytes); i++ {
		c.memory[i+512] = bytes[i]
	}

	return nil
}

// SetKeyDown will mark the specified key as down.
// Once read by the current program, the key state will be reset to up.
func (c *Chip8) SetKeyDown(index byte) {
	c.key[index] = 1
}

// GetGraphics returns the current state of the graphics memory.
// Graphics are 64x32. Each pixel is represented as a byte, 0 = off,
// !0 = on.
func (c *Chip8) GetGraphics() [64 * 32]byte {
	return c.gfx
}

// Beep returns a channel that outputs a value whenever a beep is to be played.
func (c *Chip8) Beep() <-chan struct{} {
	return c.beepOut
}

// EmulateCycle will execute a single clock cycle on this CHIP-8 cpu.
// Every cycle will return a Result containing information about the state before
// and after this cycle.
// Result will be populated regardless of whether or not an error is returned.
func (c *Chip8) EmulateCycle() (Result, error) {
	// Fetch Opcode
	opcode := uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])

	before := c.currentState()

	// Decode and Handle Opcode
	handler, ok := c.opcodes[opcode&0xF000]
	if !ok {
		return Result{
			Opcode: opcode,
			Before: before,
			After:  before,
		}, fmt.Errorf("unknown opcode: 0x%X", c.opcode)
	}

	result, err := handler(opcode)
	result.Opcode = opcode
	result.Before = before
	result.After = c.currentState()
	if err != nil {
		return result, err
	}

	select {
	case <-c.timerClock.C:
		// Update timers
		if c.delayTimer > 0 {
			c.delayTimer--
		}

		if c.soundTimer > 0 {
			if c.soundTimer == 1 {
				// Don't block if the beep routine isn't ready
				select {
				case c.beepOut <- struct{}{}:
				default:
				}
			}
			c.soundTimer--
		}
	default:
		// Skip the timers
	}

	return result, nil
}

// DrawFlag returns the current state of the draw flag.
// Iff true, the screen will need to be re-drawn using the values in
// GetGraphics.
// Reading the flag will reset it to false.
func (c *Chip8) DrawFlag() bool {
	flag := c.drawFlag
	c.drawFlag = false
	return flag
}

func (c *Chip8) currentState() ResultState {
	return ResultState{
		PC: c.pc,
		V:  c.V,
	}
}
