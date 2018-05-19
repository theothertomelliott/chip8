package chip8

import (
	"bufio"
	"fmt"
	"os"
)

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

	// The graphics system: The chip 8 has one instruction that draws sprite to the screen. Drawing is done in XOR mode and if a pixel is turned off as a result of drawing, the VF register is set. This is used for collision detection.
	// The graphics of the Chip 8 are black and white and the screen has a total of 2048 pixels (64 x 32). This can easily be implemented using an array that hold the pixel state (1 or 0):
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
}

func (c *Chip8) Initialize() {
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
}

func (c *Chip8) LoadGame(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	stats, statsErr := file.Stat()
	if statsErr != nil {
		return statsErr
	}

	var size int64 = stats.Size()
	bytes := make([]byte, size)

	bufr := bufio.NewReader(file)
	_, err = bufr.Read(bytes)
	if err != nil {
		return err
	}

	for i := 0; i < len(bytes); i++ {
		c.memory[i+512] = bytes[i]
	}

	return nil
}

func (c *Chip8) EmulateCycle() error {
	// Fetch Opcode
	c.opcode = uint16(c.memory[c.pc])<<8 | uint16(c.memory[c.pc+1])

	// Decode opcode
	switch c.opcode & 0xF000 {
	// Some opcodes //

	case 0x0000:
		switch c.opcode & 0x00FF {
		case 0x00E0:
			fmt.Println("disp_clear()")
			// Clear display
			c.gfx = [64 * 32]byte{}
			c.pc += 2
		case 0x00EE:
			fmt.Println("return;")
			c.sp--
			c.pc = c.stack[c.sp]

		default:
			return fmt.Errorf("unknown opcode: 0x%X (pc=0x%X)", c.opcode, c.pc)
		}

	case 0x1000:
		c.pc = c.opcode & 0x0FFF
		fmt.Printf("goto 0x%X;", c.pc)

	case 0x2000:
		c.stack[c.sp] = c.pc
		c.sp++
		c.pc = c.opcode & 0x0FFF
		fmt.Printf("*(0x%X)()\n", c.pc)

	case 0x3000:
		x := (c.opcode & 0x0F00) >> 8
		nn := byte(c.opcode & 0x00FF)
		if c.V[x] == nn {
			c.pc += 4
		} else {
			c.pc += 2
		}
		fmt.Printf("if(V%d==0x%X)\n", x, nn)

	case 0x4000:
		x := (c.opcode & 0x0F00) >> 8
		nn := byte(c.opcode & 0x00FF)
		if c.V[x] != nn {
			c.pc += 4
		} else {
			c.pc += 2
		}
		fmt.Printf("if(V%d!=0x%X)\n", x, nn)

	case 0x5000:
		x := (c.opcode & 0x0F00) >> 8
		y := (c.opcode & 0x00F0) >> 4
		if c.V[x] == c.V[y] {
			c.pc += 4
		} else {
			c.pc += 2
		}
		fmt.Printf("if(V%d==V%d)\n", x, y)

	case 0x6000:
		x := (c.opcode & 0x0F00) >> 8
		nn := byte(c.opcode & 0x00FF)
		fmt.Printf("V%d = 0x%X\n", x, nn)
		c.V[x] = nn
		c.pc += 2

	case 0x7000:
		x := (c.opcode & 0x0F00) >> 8
		nn := byte(c.opcode & 0x00FF)
		fmt.Printf("V%d += 0x%X\n", x, nn)
		c.V[x] += nn
		c.pc += 2

	case 0x8000:
		x := (c.opcode & 0x0F00) >> 8
		y := (c.opcode & 0x00F0) >> 4
		switch c.opcode & 0x000F {
		case 0x0000:
			fmt.Printf("V%d = V%d\n", x, y)
			c.V[x] = c.V[y]
			c.pc += 2
		case 0x0001:
			fmt.Printf("V%d |= V%d\n", x, y)
			c.V[x] |= c.V[y]
			c.pc += 2
		case 0x0002:
			fmt.Printf("V%d &= V%d\n", x, y)
			c.V[x] &= c.V[y]
			c.pc += 2
		case 0x0003:
			fmt.Printf("V%d ^= V%d\n", x, y)
			c.V[x] ^= c.V[y]
			c.pc += 2
		case 0x0004:
			fmt.Printf("V%d += V%d\n", x, y)
			if c.V[y] > (0xFF - c.V[x]) {
				c.V[0xF] = 1 //carry
			} else {
				c.V[0xF] = 0
			}
			c.V[x] += c.V[y]
			c.pc += 2
		case 0x0005:
			fmt.Printf("V%d -= V%d\n", x, y)
			if c.V[y] > c.V[x] {
				c.V[0xF] = 0 //borrow
			} else {
				c.V[0xF] = 1
			}
			c.V[x] -= c.V[y]
			c.pc += 2
		case 0x0006:
			fmt.Printf("V%d=V%d=V%d>>1\n", x, y, y)
			// TODO:
			// Shifts VY right by one and stores the result to VX (VY remains unchanged).
			// VF is set to the value of the least significant bit of VY before the shift.[2]
			c.pc += 2
		case 0x0007:
			fmt.Printf("V%d=V%d-V%d\n", x, y, x)
			// TODO:
			// Sets VX to VY minus VX. VF is set to 0 when there's a borrow, and 1 when there isn't.
			c.pc += 2
		case 0x000E:
			fmt.Printf("V%d=V%d=V%d<<1\n", x, y, y)
			// TODO:
			// Shifts VY left by one and copies the result to VX.
			// VF is set to the value of the most significant bit of VY before the shift.
			c.pc += 2
		default:
			return fmt.Errorf("unknown opcode: 0x%X (pc=0x%X)", c.opcode, c.pc)
		}

	case 0x9000:
		x := (c.opcode & 0x0F00) >> 8
		y := (c.opcode & 0x00F0) >> 4
		if c.V[x] != c.V[y] {
			c.pc += 4
		} else {
			c.pc += 2
		}
		fmt.Printf("if(V%d!=V%d)\n", x, y)

	case 0xA000: // ANNN: Sets I to the address NNN
		c.I = c.opcode & 0x0FFF
		fmt.Printf("I = 0x%X\n", c.I)
		c.pc += 2

	case 0xB000:
		nnn := c.opcode & 0x0FFF
		fmt.Printf("PC=V0+0x%X\n", nnn)
		c.pc = uint16(c.V[0]) + nnn

	case 0xC000:
		x := uint16(c.V[(c.opcode&0x0F00)>>8])
		nn := c.opcode & 0x00FF
		fmt.Printf("V%d=rand()&0x%X\n", x, nn)
		// TODO:
		// Sets VX to the result of a bitwise and operation on a random number (Typically: 0 to 255) and NN.
		c.pc += 2

	case 0xD000:
		x := uint16(c.V[(c.opcode&0x0F00)>>8])
		y := uint16(c.V[(c.opcode&0x00F0)>>4])
		height := c.opcode & 0x000F
		var pixel uint16

		fmt.Printf("draw(V%d,V%d,%d)\n", x, y, height)

		c.V[0xF] = 0
		for yline := uint16(0); yline < height; yline++ {
			pixel = uint16(c.memory[c.I+yline])
			for xline := uint16(0); xline < 8; xline++ {
				if (pixel & (0x80 >> xline)) != 0 {
					if c.gfx[(x+xline+((y+yline)*64))] == 1 {
						c.V[0xF] = 1
						c.gfx[x+xline+((y+yline)*64)] ^= 1
					}
				}
			}
		}

		c.drawFlag = true
		c.pc += 2

	case 0xE000:
		x := (c.opcode & 0x0F00) >> 8
		switch c.opcode & 0x00FF {
		case 0x009E:
			// TODO:
			// Skips the next instruction if the key stored in VX is pressed.
			// (Usually the next instruction is a jump to skip a code block)
			fmt.Printf("if(key()==V%d)\n", x)
		case 0x00A1:
			// TODO:
			// Skips the next instruction if the key stored in VX isn't pressed.
			// (Usually the next instruction is a jump to skip a code block)
			fmt.Printf("if(key()!=V%d)\n", x)
		}

	case 0xF000:
		x := (c.opcode & 0x0F00) >> 8
		switch c.opcode & 0x00FF {
		case 0x0007:
			// TODO:
			fmt.Println("Vx = get_delay()")
		case 0x000A:
			// TODO:
			fmt.Println("Vx = get_key()")
		case 0x0015:
			// TODO:
			fmt.Println("delay_timer(Vx)")
		case 0x0018:
			// TODO:
			fmt.Println("sound_timer(Vx)")
		case 0x001E:
			// TODO:
			fmt.Println("I +=Vx")
		case 0x0029:
			// Sets I to the location of the sprite for the character in VX. Characters 0-F (in hexadecimal) are represented by a 4x5 font.
			c.I = uint16(c.V[x]) * 5
			c.pc += 2
			fmt.Printf("I=sprite_addr[V%d]\n", x)
		case 0x0033:
			c.memory[c.I] = c.V[x] / 100
			c.memory[c.I+1] = (c.V[x] / 10) % 10
			c.memory[c.I+2] = (c.V[x] % 100) % 10
			c.pc += 2
			fmt.Printf("set_BCD(Vx);\n")
			fmt.Printf("*(I + 0) = BCD(3)\n")
			fmt.Printf("*(I + 1) = BCD(2)\n")
			fmt.Printf("*(I + 2) = BCD(1)\n")
		case 0x0055:
			// TODO:
			fmt.Println("reg_dump(Vx, &I)")
			c.pc += 2
		case 0x0065:
			for i := uint16(0); i <= x; i++ {
				c.V[i] = c.memory[c.I+i]
			}
			c.pc += 2
			fmt.Printf("reg_load(V%d,&I)", x)
		default:
			return fmt.Errorf("unknown opcode: 0x%X (pc=0x%X)", c.opcode, c.pc)
		}

	default:
		return fmt.Errorf("unknown opcode: 0x%X (pc=0x%X)", c.opcode, c.pc)
	}

	// Update timers
	if c.delayTimer > 0 {
		c.delayTimer--
	}

	if c.soundTimer > 0 {
		if c.soundTimer == 1 {
			fmt.Println("BEEP!")
			c.soundTimer--
		}
	}
	return nil
}

func (c *Chip8) DrawFlag() bool {
	return c.drawFlag
}

func (c *Chip8) SetKeys() {

}
