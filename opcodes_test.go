package chip8

import "testing"

func Test0x00E0(t *testing.T) {
	cpu := initCPU()
	// Fill graphics
	for i := 0; i < len(cpu.gfx); i++ {
		cpu.gfx[i] = 1
	}

	r, err := cpu.opcode0x0000(0x00E0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectOpcodeType(t, r, "0x00E0")

	// Fill graphics
	for i := 0; i < len(cpu.gfx); i++ {
		if cpu.gfx[i] != 0 {
			t.Errorf("Graphics were not cleared as expected")
			return
		}
	}
}

func Test0x00EE(t *testing.T) {
	cpu := initCPU()
	cpu.sp = 0
	cpu.stack[0] = 0x321

	r, err := cpu.opcode0x0000(0x00EE)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectOpcodeType(t, r, "0x00EE")
	// Previous opcode + 2
	expectPC(t, cpu, 0x321+2)
}

func Test0x1NNN(t *testing.T) {
	cpu := initCPU()
	r, err := cpu.opcode0x1000(0x1123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectOpcodeType(t, r, "0x1NNN")
	expectPC(t, cpu, 0x123)
}

func Test0x2NNN(t *testing.T) {
	cpu := initCPU()
	cpu.pc = 0x123
	r, err := cpu.opcode0x2000(0x2321)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectOpcodeType(t, r, "0x2NNN")
	expectPC(t, cpu, 0x321)
	expectStack(t, cpu, 0x123)
}

func Test0x3XNN(t *testing.T) {
	var tests = []struct {
		name       string
		v3         byte
		expectedPC uint16
	}{
		{
			name:       "equal",
			v3:         0x001,
			expectedPC: 0x4,
		},
		{
			name:       "not equal",
			v3:         0x002,
			expectedPC: 0x2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu := initCPU()
			cpu.pc = 0x0
			cpu.V[3] = test.v3
			r, err := cpu.opcode0x3000(0x3301)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expectOpcodeType(t, r, "0x3XNN")
			expectPC(t, cpu, test.expectedPC)
		})
	}
}

func Test0x4XNN(t *testing.T) {
	var tests = []struct {
		name       string
		v4         byte
		expectedPC uint16
	}{
		{
			name:       "equal",
			v4:         0x001,
			expectedPC: 0x2,
		},
		{
			name:       "not equal",
			v4:         0x002,
			expectedPC: 0x4,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu := initCPU()
			cpu.pc = 0x0
			cpu.V[4] = test.v4
			r, err := cpu.opcode0x4000(0x4401)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expectOpcodeType(t, r, "0x4XNN")
			expectPC(t, cpu, test.expectedPC)
		})
	}
}

func Test0x5XY0(t *testing.T) {
	var tests = []struct {
		name       string
		v5         byte
		v6         byte
		expectedPC uint16
	}{
		{
			name:       "equal",
			v5:         0x001,
			v6:         0x001,
			expectedPC: 0x4,
		},
		{
			name:       "not equal",
			v5:         0x002,
			v6:         0x003,
			expectedPC: 0x2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu := initCPU()
			cpu.pc = 0x0
			cpu.V[5] = test.v5
			cpu.V[6] = test.v6
			r, err := cpu.opcode0x5000(0x5560)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expectOpcodeType(t, r, "0x5XY0")
			expectPC(t, cpu, test.expectedPC)
		})
	}
}

func Test0x6XNN(t *testing.T) {
	cpu := initCPU()
	r, err := cpu.opcode0x6000(0x6123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectOpcodeType(t, r, "0x6XNN")
	expectRegister(t, cpu, 1, 0x23)
}

func Test0x7XNN(t *testing.T) {
	cpu := initCPU()
	cpu.V[7] = 5
	r, err := cpu.opcode0x7000(0x7723)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectOpcodeType(t, r, "0x7XNN")
	expectRegister(t, cpu, 7, 0x28)
}

func Test0x8XY3(t *testing.T) {
	cpu := initCPU()
	cpu.V[0] = 0x0F
	cpu.V[1] = 0x1F
	r, err := cpu.opcode0x8000(0x8013)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectOpcodeType(t, r, "0x8XY3")

	expectRegister(t, cpu, 0, 0x10)
	expectRegister(t, cpu, 1, 0x1F)
}

func Test0x8XY4(t *testing.T) {
	var tests = []struct {
		name       string
		v0         byte
		v1         byte
		expectedV0 byte
		expectedV1 byte
		expectedVF byte
	}{
		{
			name:       "no carry",
			v0:         0x05,
			v1:         0x05,
			expectedV0: 0x0A,
			expectedV1: 0x05,
			expectedVF: 0,
		},
		{
			name:       "carry",
			v0:         0xFF,
			v1:         0x01,
			expectedV0: 0x00,
			expectedV1: 0x01,
			expectedVF: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu := initCPU()
			cpu.V[0] = test.v0
			cpu.V[1] = test.v1
			r, err := cpu.opcode0x8000(0x8014)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expectOpcodeType(t, r, "0x8XY4")

			expectRegister(t, cpu, 0, test.expectedV0)
			expectRegister(t, cpu, 1, test.expectedV1)
			expectRegister(t, cpu, 0xF, test.expectedVF)
		})
	}
}

func Test0x8XY5(t *testing.T) {
	var tests = []struct {
		name       string
		v0         byte
		v1         byte
		expectedV0 byte
		expectedV1 byte
		expectedVF byte
	}{
		{
			name:       "no borrow",
			v0:         0x05,
			v1:         0x02,
			expectedV0: 0x03,
			expectedV1: 0x02,
			expectedVF: 1,
		},
		{
			name:       "borrow",
			v0:         0x01,
			v1:         0x02,
			expectedV0: 0xFF,
			expectedV1: 0x02,
			expectedVF: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu := initCPU()
			cpu.V[0] = test.v0
			cpu.V[1] = test.v1
			r, err := cpu.opcode0x8000(0x8015)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expectOpcodeType(t, r, "0x8XY5")

			expectRegister(t, cpu, 0, test.expectedV0)
			expectRegister(t, cpu, 1, test.expectedV1)
			expectRegister(t, cpu, 0xF, test.expectedVF)
		})
	}
}

func Test0x8XY6(t *testing.T) {
	var tests = []struct {
		name       string
		v0         byte
		v1         byte
		expectedV0 byte
		expectedV1 byte
		expectedVF byte
	}{
		{
			name:       "least significant bit of 0",
			v0:         0x00,
			v1:         0x02,
			expectedV0: 0x01,
			expectedV1: 0x02,
			expectedVF: 0,
		},
		{
			name:       "least significant bit of 1",
			v0:         0x00,
			v1:         0x03,
			expectedV0: 0x01,
			expectedV1: 0x03,
			expectedVF: 1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cpu := initCPU()
			cpu.V[0] = test.v0
			cpu.V[1] = test.v1
			r, err := cpu.opcode0x8000(0x8016)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expectOpcodeType(t, r, "0x8XY6")

			expectRegister(t, cpu, 0, test.expectedV0)
			expectRegister(t, cpu, 1, test.expectedV1)
			expectRegister(t, cpu, 0xF, test.expectedVF)
		})
	}
}

func Test0xFX18(t *testing.T) {
	cpu := initCPU()
	cpu.V[0] = 0x0F
	r, err := cpu.opcode0xF000(0xF018)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expectOpcodeType(t, r, "0xFX18")
	if cpu.soundTimer != 0x0F {
		t.Errorf("Expected sountTimer to be 0x0F, got 0x%X", cpu.soundTimer)
	}
}

func initCPU() *Chip8 {
	cpu := &Chip8{}
	cpu.initialize()
	return cpu
}

func expectRegister(t *testing.T, cpu *Chip8, register int, expected byte) {
	t.Helper()
	if cpu.V[register] != expected {
		t.Errorf("V%d should be 0x%X, got 0x%X", register, expected, cpu.V[register])
	}
}

func expectPC(t *testing.T, cpu *Chip8, expected uint16) {
	t.Helper()
	if cpu.pc != expected {
		t.Errorf("PC should be 0x%X, got 0x%X", expected, cpu.pc)
	}
}

// expectStack tests for a particular value on the top of the stack
func expectStack(t *testing.T, cpu *Chip8, expected uint16) {
	t.Helper()
	if cpu.stack[cpu.sp] != expected {
		t.Errorf("Top of stack should be 0x%X, got 0x%X", expected, cpu.stack[cpu.sp])
	}
}

func expectOpcodeType(t *testing.T, r Result, expected string) {
	if r.OpcodeType != expected {
		t.Logf("%+v", r)
		t.Errorf("Expected opcode type %s, got %s", expected, r.OpcodeType)
	}
}
