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

func expectOpcodeType(t *testing.T, r Result, expected string) {
	if r.OpcodeType != expected {
		t.Logf("%+v", r)
		t.Errorf("Expected opcode type %s, got %s", expected, r.OpcodeType)
	}
}
