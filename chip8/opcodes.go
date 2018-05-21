package chip8

import (
	"fmt"
	"math/rand"
)

type opcodeHandler func(opcode uint16) (Result, error)

func (c *Chip8) registerOpcodeHandlers() {
	c.opcodes = map[uint16]opcodeHandler{
		0x0000: c.opcode0x0000,
		0x1000: c.opcode0x1000,
		0x2000: c.opcode0x2000,
		0x3000: c.opcode0x3000,
		0x4000: c.opcode0x4000,
		0x5000: c.opcode0x5000,
		0x6000: c.opcode0x6000,
		0x7000: c.opcode0x7000,
		0x8000: c.opcode0x8000,
		0x9000: c.opcode0x9000,
		0xA000: c.opcode0xA000,
		0xB000: c.opcode0xB000,
		0xC000: c.opcode0xC000,
		0xD000: c.opcode0xD000,
		0xE000: c.opcode0xE000,
		0xF000: c.opcode0xF000,
	}
}

func (c *Chip8) opcode0x0000(opcode uint16) (Result, error) {
	result := Result{}

	switch opcode & 0x00FF {
	case 0x00E0:
		result.Pseudo = fmt.Sprint("disp_clear()")
		// Clear display
		c.gfx = [64 * 32]byte{}
		c.pc += 2
	case 0x00EE:
		result.Pseudo = fmt.Sprint("return;")
		c.sp--
		c.pc = c.stack[c.sp] + 2

	default:
		return result, fmt.Errorf("unknown opcode: 0x%X", opcode)
	}
	return result, nil
}

func (c *Chip8) opcode0x1000(opcode uint16) (Result, error) {
	c.pc = opcode & 0x0FFF
	return Result{
		Pseudo: fmt.Sprintf("goto 0x%X;", c.pc),
	}, nil
}

func (c *Chip8) opcode0x2000(opcode uint16) (Result, error) {
	c.stack[c.sp] = c.pc
	c.sp++
	c.pc = opcode & 0x0FFF
	return Result{
		Pseudo: fmt.Sprintf("*(0x%X)()", c.pc),
	}, nil
}

func (c *Chip8) opcode0x3000(opcode uint16) (Result, error) {
	x := (opcode & 0x0F00) >> 8
	nn := byte(opcode & 0x00FF)
	if c.V[x] == nn {
		c.pc += 4
	} else {
		c.pc += 2
	}
	return Result{
		Pseudo: fmt.Sprintf("if(V%d==0x%X)", x, nn),
	}, nil
}

func (c *Chip8) opcode0x4000(opcode uint16) (Result, error) {
	x := (opcode & 0x0F00) >> 8
	nn := byte(opcode & 0x00FF)
	if c.V[x] != nn {
		c.pc += 4
	} else {
		c.pc += 2
	}
	return Result{
		Pseudo: fmt.Sprintf("if(V%d!=0x%X)", x, nn),
	}, nil
}

func (c *Chip8) opcode0x5000(opcode uint16) (Result, error) {
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	if c.V[x] == c.V[y] {
		c.pc += 4
	} else {
		c.pc += 2
	}
	return Result{
		Pseudo: fmt.Sprintf("if(V%d==V%d)", x, y),
	}, nil
}

func (c *Chip8) opcode0x6000(opcode uint16) (Result, error) {
	x := (opcode & 0x0F00) >> 8
	nn := byte(opcode & 0x00FF)
	c.V[x] = nn
	c.pc += 2
	return Result{
		Pseudo: fmt.Sprintf("V%d = 0x%X", x, nn),
	}, nil
}

func (c *Chip8) opcode0x7000(opcode uint16) (Result, error) {
	x := (opcode & 0x0F00) >> 8
	nn := byte(opcode & 0x00FF)
	c.V[x] += nn
	c.pc += 2
	return Result{
		Pseudo: fmt.Sprintf("V%d += 0x%X", x, nn),
	}, nil
}

func (c *Chip8) opcode0x8000(opcode uint16) (Result, error) {
	result := Result{}
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	switch opcode & 0x000F {
	case 0x0000:
		c.V[x] = c.V[y]
		c.pc += 2
		result.Pseudo = fmt.Sprintf("V%d = V%d", x, y)
	case 0x0001:
		c.V[x] |= c.V[y]
		c.pc += 2
		result.Pseudo = fmt.Sprintf("V%d |= V%d", x, y)
	case 0x0002:
		c.V[x] &= c.V[y]
		c.pc += 2
		result.Pseudo = fmt.Sprintf("V%d &= V%d", x, y)
	case 0x0003:
		c.V[x] ^= c.V[y]
		c.pc += 2
		result.Pseudo = fmt.Sprintf("V%d ^= V%d", x, y)
	case 0x0004:
		if c.V[y] > (0xFF - c.V[x]) {
			c.V[0xF] = 1 //carry
		} else {
			c.V[0xF] = 0
		}
		c.V[x] += c.V[y]
		c.pc += 2
		result.Pseudo = fmt.Sprintf("V%d += V%d", x, y)
	case 0x0005:
		if c.V[y] > c.V[x] {
			c.V[0xF] = 0 //borrow
		} else {
			c.V[0xF] = 1
		}
		c.V[x] -= c.V[y]
		c.pc += 2
		result.Pseudo = fmt.Sprintf("V%d -= V%d", x, y)
	case 0x0006:
		c.V[x] = c.V[y] >> 1
		c.V[0xF] = c.V[y] & 0x01
		c.pc += 2
		result.Pseudo = fmt.Sprintf("V%d=V%d=V%d>>1", x, y, y)
	case 0x0007:
		if c.V[x] > c.V[y] {
			c.V[0xF] = 0 //borrow
		} else {
			c.V[0xF] = 1
		}
		c.V[x] = c.V[y] - c.V[x]
		c.pc += 2
		result.Pseudo = fmt.Sprintf("V%d=V%d-V%d", x, y, x)
	case 0x000E:
		c.V[x] = c.V[y] << 1
		c.V[0xF] = c.V[y] & 0x80
		c.pc += 2
		result.Pseudo = fmt.Sprintf("V%d=V%d=V%d<<1", x, y, y)
	default:
		return Result{}, fmt.Errorf("unknown opcode: 0x%X", opcode)
	}
	return result, nil
}

func (c *Chip8) opcode0x9000(opcode uint16) (Result, error) {
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	if c.V[x] != c.V[y] {
		c.pc += 4
	} else {
		c.pc += 2
	}
	return Result{
		Pseudo: fmt.Sprintf("if(V%d!=V%d)", x, y),
	}, nil
}

func (c *Chip8) opcode0xA000(opcode uint16) (Result, error) {
	c.I = opcode & 0x0FFF
	c.pc += 2
	return Result{
		Pseudo: fmt.Sprintf("I = 0x%X", c.I),
	}, nil
}

func (c *Chip8) opcode0xB000(opcode uint16) (Result, error) {
	nnn := opcode & 0x0FFF
	c.pc = uint16(c.V[0]) + nnn
	return Result{
		Pseudo: fmt.Sprintf("PC=V0+0x%X", nnn),
	}, nil
}

func (c *Chip8) opcode0xC000(opcode uint16) (Result, error) {
	x := uint16(opcode&0x0F00) >> 8
	nn := opcode & 0x00FF
	c.V[x] = byte(rand.Float32()*255) & byte(nn)
	c.pc += 2
	return Result{
		Pseudo: fmt.Sprintf("V%d=rand()&0x%X", x, nn),
	}, nil
}

func (c *Chip8) opcode0xD000(opcode uint16) (Result, error) {
	x := uint16(c.V[(opcode&0x0F00)>>8])
	y := uint16(c.V[(opcode&0x00F0)>>4])
	height := opcode & 0x000F
	var pixel uint16

	c.V[0xF] = 0
	for yline := uint16(0); yline < height; yline++ {
		pixel = uint16(c.memory[c.I+yline])
		for xline := uint16(0); xline < 8; xline++ {
			index := (x + xline + ((y + yline) * 64))
			if index > uint16(len(c.gfx)) {
				continue
			}
			if (pixel & (0x80 >> xline)) != 0 {
				if c.gfx[index] == 1 {
					c.V[0xF] = 1
				}
				c.gfx[index] ^= 1
			}
		}
	}

	c.drawFlag = true
	c.pc += 2

	return Result{
		Pseudo: fmt.Sprintf("draw(V%d,V%d,%d)", x, y, height),
	}, nil
}

func (c *Chip8) opcode0xE000(opcode uint16) (Result, error) {
	result := Result{}
	x := (opcode & 0x0F00) >> 8
	switch opcode & 0x00FF {
	case 0x009E:
		if c.key[c.V[x]] != 0 {
			c.pc += 4
			c.key[c.V[x]] = 0
		} else {
			c.pc += 2
		}
		result.Pseudo = fmt.Sprintf("if(key()==V%d)", x)
	case 0x00A1:
		if c.key[c.V[x]] == 0 {
			c.pc += 4
		} else {
			c.key[c.V[x]] = 0
			c.pc += 2
		}
		result.Pseudo = fmt.Sprintf("if(key()!=V%d)", x)
	}
	return result, nil
}

func (c *Chip8) opcode0xF000(opcode uint16) (Result, error) {
	result := Result{}
	x := (opcode & 0x0F00) >> 8
	switch opcode & 0x00FF {
	case 0x0007:
		c.V[x] = c.delayTimer
		c.pc += 2
		result.Pseudo = fmt.Sprint("Vx = get_delay()")
	case 0x000A:
		for index, k := range c.key {
			if k != 0 {
				c.V[x] = byte(index)
				c.pc += 2
				break
			}
		}
		c.key[c.V[x]] = 0
		result.Pseudo = fmt.Sprint("Vx = get_key()")
	case 0x0015:
		c.delayTimer = c.V[x]
		c.pc += 2
		result.Pseudo = fmt.Sprintf("delay_timer(V%d)", x)

	case 0x0018:
		c.soundTimer = c.V[x]
		c.pc += 2
		result.Pseudo = fmt.Sprintf("sound_timer(V%d)", x)

	case 0x001E:
		c.I += uint16(c.V[x])
		c.pc += 2
		result.Pseudo = fmt.Sprintf("I += V%d", x)

	case 0x0029:
		// Sets I to the location of the sprite for the character in VX. Characters 0-F (in hexadecimal) are represented by a 4x5 font.
		c.I = uint16(c.V[x]) * 5
		c.pc += 2
		result.Pseudo = fmt.Sprintf("I=sprite_addr[V%d]", x)
	case 0x0033:
		c.memory[c.I] = c.V[x] / 100
		c.memory[c.I+1] = (c.V[x] / 10) % 10
		c.memory[c.I+2] = (c.V[x] % 100) % 10
		c.pc += 2
		result.Pseudo = fmt.Sprintf("set_BCD(V%d);\n*(I + 0) = BCD(3)\n*(I + 1) = BCD(2)\n*(I + 2) = BCD(1)", x)
	case 0x0055:
		for i := uint16(0); i <= x; i++ {
			c.memory[c.I+i] = c.V[i]
		}
		c.pc += 2
		result.Pseudo = fmt.Sprintf("reg_dump(V%d, &I)", x)
	case 0x0065:
		for i := uint16(0); i <= x; i++ {
			c.V[i] = c.memory[c.I+i]
		}
		c.pc += 2
		result.Pseudo = fmt.Sprintf("reg_load(V%d,&I)", x)
	default:
		return Result{}, fmt.Errorf("unknown opcode: 0x%X", opcode)
	}

	return result, nil
}
