package cpu

import (
	"log"
	"strconv"

	decoder "github.com/giammirove/gbemu/libs/decoder"
	"github.com/giammirove/gbemu/libs/interrupts"
	"github.com/giammirove/gbemu/libs/mmu"
	registers "github.com/giammirove/gbemu/libs/registers"
	"github.com/giammirove/gbemu/libs/timer"
	"github.com/giammirove/gbemu/libs/utility"
)

const _RST_BASE = 0xC7
const _JOYPAD = 0xFF00

func getValueByOperand(op decoder.Operand_t) uint {

	value := uint(0)
	if op.Withvalue {
		value = op.Value
		if !op.Immediate {
			Cycle(4)
			value = mmu.ReadFromMemory(op.Value, op.Bytes)
		}
	} else {
		value = registers.Get(op.Name)()
		if !op.Immediate {
			Cycle(4)
			value = mmu.ReadFromMemory(value, op.Bytes)
		}
	}

	return value
}

func setValueByOperand(op decoder.Operand_t, val uint) {

	if op.Withvalue {
		if !op.Immediate {
			Cycle(4)
			mmu.WriteToMemory(op.Value, val)
		} else {
			log.Fatal("Could this really happend?")
		}
	} else {
		if op.Immediate {
			registers.Set(op.Name)(uint(val))
		} else {
			Cycle(4)
			mmu.WriteToMemory(registers.Get(op.Name)(), uint(val), op.Bytes)
		}
	}
}
func notHandled(instruction decoder.Instruction_t) {
	log.Fatalf("%s opcode not recognized %02X\n", instruction.Mnemonic, instruction.Opcode)
}

func handlePUSH(instruction decoder.Instruction_t) {
	StackPush(registers.Get(instruction.Operands[0].Name)(), 2)
	Cycle(4)
}

func handlePOP(instruction decoder.Instruction_t) {
	value := StackPOP()
	registers.Set(instruction.Operands[0].Name)(value)
	if instruction.Operands[0].Name == "AF" {
		// reset last 4 bit
		registers.SetAF(value & 0xFFF0)
	}
}

func handleLD(instruction decoder.Instruction_t) {
	switch instruction.Opcode {
	case 0x06, 0xE, 0x16, 0x1E, 0x26, 0x2E:
		handleLD_nn_n(instruction)
		break
	case 0x7F, 0x78, 0x79, 0x7A, 0x7B, 0x7C, 0x7D, 0x7E, 0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4A, 0x4B, 0x4C, 0x4D, 0x4E, 0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5A, 0x5B, 0x5C, 0x5D, 0x5E, 0x5F, 0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6A, 0x6B, 0x6C, 0x6D, 0x6E, 0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x36:
		handleLD_r1_r2(instruction)
		break
	case 0x0A, 0x1A, 0xFA, 0x3E:
		handleLD_A_n(instruction)
		break
	case 0x4F, 0x6F, 0x02, 0x12, 0x77, 0xEA:
		handleLD_n_A(instruction)
		break
	case 0xF2:
		handleLD_A_C(instruction)
		break
	case 0xE2:
		handleLD_C_A(instruction)
		break
	case 0x3A:
		handleLDD_A_HL(instruction)
		break
	case 0x32:
		handleLDD_HL_A(instruction)
		break
	case 0x2A:
		handleLDI_A_HL(instruction)
		break
	case 0x22:
		handleLDI_HL_A(instruction)
		break
	case 0xE0:
		handleLDH_n_A(instruction)
		break
	case 0xF0:
		handleLDH_A_n(instruction)
		break
	case 0x01, 0x11, 0x21, 0x31:
		handleLD_n_nn(instruction)
		break
	case 0xF9:
		handleLD_SP_HL(instruction)
		break
	case 0xF8:
		handleLDHL_SP_n(instruction)
		break
	case 0x08:
		handleLD_nn_SP(instruction)
		break
	default:
		notHandled(instruction)
	}

}

func handleLD_nn_n(instruction decoder.Instruction_t) {
	registers.Set(instruction.Operands[0].Name)(instruction.Operands[1].Value)
}
func handleLD_r1_r2(instructon decoder.Instruction_t) {
	value := getValueByOperand(instructon.Operands[1])
	setValueByOperand(instructon.Operands[0], value)
}
func handleLD_A_n(instruction decoder.Instruction_t) {
	value := getValueByOperand(instruction.Operands[1])
	registers.SetA(value)
}
func handleLD_n_A(instruction decoder.Instruction_t) {
	setValueByOperand(instruction.Operands[0], registers.A())
}
func handleLD_A_C(instruction decoder.Instruction_t) {
	Cycle(4)
	registers.SetA(mmu.ReadFromMemory(_JOYPAD + registers.C()))
}
func handleLD_C_A(instruction decoder.Instruction_t) {
	Cycle(4)
	mmu.WriteToMemory(_JOYPAD+registers.C(), registers.A())
}
func handleLDD_A_HL(instruction decoder.Instruction_t) {
	Cycle(4)
	registers.SetA(mmu.ReadFromMemory(registers.HL(), 2))
	registers.SetHL(registers.HL() - 1)
}
func handleLDD_HL_A(instruction decoder.Instruction_t) {
	Cycle(4)
	mmu.WriteToMemory(registers.HL(), registers.A())
	registers.SetHL(registers.HL() - 1)
}
func handleLDI_A_HL(instruction decoder.Instruction_t) {
	Cycle(4)
	registers.SetA(mmu.ReadFromMemory(registers.HL(), 2))
	registers.SetHL(registers.HL() + 1)
}
func handleLDI_HL_A(instruction decoder.Instruction_t) {
	Cycle(4)
	mmu.WriteToMemory(registers.HL(), registers.A())
	registers.SetHL(registers.HL() + 1)
}
func handleLDH_n_A(instruction decoder.Instruction_t) {
	Cycle(4)
	mmu.WriteToMemory(_JOYPAD+instruction.Operands[0].Value, registers.A())
}
func handleLDH_A_n(instruction decoder.Instruction_t) {
	Cycle(4)
	registers.SetA(mmu.ReadFromMemory(_JOYPAD + instruction.Operands[1].Value))
}
func handleLD_n_nn(instruction decoder.Instruction_t) {
	registers.Set(instruction.Operands[0].Name)(instruction.Operands[1].Value)
}
func handleLD_SP_HL(instruction decoder.Instruction_t) {
	Cycle(4)
	registers.SetSP(registers.HL())
}
func handleLDHL_SP_n(instruction decoder.Instruction_t) {
	Cycle(4)
	// value is signed
	n := int8(instruction.Operands[2].Value)
	value := uint(int(n) + int(registers.SP()))
	registers.SetHL(value)
	// update flags
	registers.SetZeroFlag(false)
	registers.SetSubstractionFlag(false)
	registers.SetHalfCarryFlag(utility.CheckHalfCarry(registers.SP(), uint(n)))
	registers.SetCarryFlag(utility.CheckCarry(registers.SP(), uint(n)))
}
func handleLD_nn_SP(instruction decoder.Instruction_t) {
	Cycle(8)
	mmu.WriteToMemory(instruction.Operands[0].Value, registers.SP(), 2)
}

func handleJP(instruction decoder.Instruction_t) {
	switch instruction.Opcode {
	case 0xC3:
		handleJP_nn(instruction)
		break
	case 0xC2, 0xCA, 0xD2, 0xDA:
		handleJP_cc_nn(instruction)
		break
	case 0xE9:
		handleJP_HL(instruction)
		break
	case 0x18:
		handleJR_n(instruction)
		break
	case 0x20, 0x28, 0x30, 0x38:
		handleJR_cc_n(instruction)
		break
	default:
		notHandled(instruction)
	}
}
func handleJP_nn(instruction decoder.Instruction_t) {
	Cycle(4)
	registers.SetPC(instruction.Operands[0].Value)
}
func handleJP_cc_nn(instruction decoder.Instruction_t) {
	if registers.GetFlag(instruction.Operands[0].Name) {
		// jump costs clock Cycles
		Cycle(4)
		registers.SetPC(instruction.Operands[1].Value)
	}
}
func handleJP_HL(instruction decoder.Instruction_t) {
	registers.SetPC(registers.HL())
}
func handleJR_n(instruction decoder.Instruction_t) {
	Cycle(4)
	// value is signed
	registers.SetPC(uint(int(registers.PC()) + int(int8(instruction.Operands[0].Value))))
}
func handleJR_cc_n(instruction decoder.Instruction_t) {
	if registers.GetFlag(instruction.Operands[0].Name) {
		// jump costs clock Cycles
		Cycle(4)
		// value is signed
		registers.SetPC(uint(int(registers.PC()) + int(int8(instruction.Operands[1].Value))))
	}
}

func handleCALL(instruction decoder.Instruction_t) {
	switch instruction.Opcode {
	case 0xCD:
		handleCALL_nn(instruction)
		break
	case 0xC4, 0xCC, 0xD4, 0xDC:
		handleCALL_cc_nn(instruction)
		break
	default:
		notHandled(instruction)
	}
}
func handleCALL_nn(instruction decoder.Instruction_t) {
	Cycle(4)
	StackPush(registers.PC(), 2)
	registers.SetPC(instruction.Operands[0].Value)
}
func handleCALL_cc_nn(instruction decoder.Instruction_t) {
	if registers.GetFlag(instruction.Operands[0].Name) {
		// jump costs clock Cycles
		Cycle(4)
		StackPush(registers.PC(), 2)
		registers.SetPC(instruction.Operands[1].Value)
	}
}

func handleRST(instruction decoder.Instruction_t) {
	Cycle(4)
	value := instruction.Opcode - _RST_BASE + 0x0000
	StackPush(registers.PC(), 2)
	registers.SetPC(value)
}

func handleRET(instruction decoder.Instruction_t) {
	ret := true
	if len(instruction.Operands) == 1 {
		ret = registers.GetFlag(instruction.Operands[0].Name)
	}
	if ret {
		// jump costs clock Cycles
		if len(instruction.Operands) == 1 {
			Cycle(4)
		}
		value := StackPOP()
		registers.SetPC(value)
	}
	Cycle(4)
}
func handleRETI(instruction decoder.Instruction_t) {
	value := StackPOP()
	registers.SetPC(value)
	interrupts.EnableInterrupts()
	Cycle(4)
}

func handleDI(instruction decoder.Instruction_t) {
	interrupts.DisableInterrupts()
}
func handleEI(instruction decoder.Instruction_t) {
	interrupts.EnableInterrupts()
}

func handleSWAP(instruction decoder.Instruction_t) {
	value := getValueByOperand(instruction.Operands[0])
	res := utility.Swap8bit(uint8(value))
	setValueByOperand(instruction.Operands[0], uint(res))

	registers.SetFlags(uint8(res) == 0, false, false, false)
}

func handleDAA(instruction decoder.Instruction_t) {
	u := uint8(0)
	fc := false

	if registers.H_flag() || (!registers.N_flag() && (registers.A()&0xF) > 9) {
		u = 6
	}

	if registers.C_flag() || (!registers.N_flag() && registers.A() > 0x99) {
		u |= 0x60
		fc = true
	}

	if registers.N_flag() {
		registers.SetA(uint(uint8(registers.A()) - uint8(u)))
	} else {
		registers.SetA(uint(uint8(registers.A()) + uint8(u)))
	}

	registers.SetZeroFlag(uint8(registers.A()) == 0)
	registers.SetHalfCarryFlag(false)
	registers.SetCarryFlag(fc)

}

func handleCPL(instruction decoder.Instruction_t) {
	registers.SetA(uint(^uint8(registers.A())))
	registers.SetSubstractionFlag(true)
	registers.SetHalfCarryFlag(true)
}
func handleCCF(instrucion decoder.Instruction_t) {
	registers.SetSubstractionFlag(false)
	registers.SetHalfCarryFlag(false)
	registers.SetCarryFlag(!registers.C_flag())
}
func handleSCF(instruction decoder.Instruction_t) {
	registers.SetSubstractionFlag(false)
	registers.SetHalfCarryFlag(false)
	registers.SetCarryFlag(true)
}
func handleNOP(instruction decoder.Instruction_t) {
}
func handleHALT(instruction decoder.Instruction_t) {
	SetHalted(true)
}
func handleSTOP(instruction decoder.Instruction_t) {
	timer.ResetDIV()
}

func handleRLCA(instruction decoder.Instruction_t) {
	// 11111110 -> 11111101
	a := uint8(registers.A())
	r := (a >> 7) & 1
	a = (a << 1) | r
	registers.SetA(uint(a))

	c := false
	if r == 1 {
		c = true
	}

	registers.SetFlags(false, false, false, c)
}
func handleRLA(instruction decoder.Instruction_t) {
	a := uint8(registers.A())
	r := (a >> 7) & 1
	cc := uint8(0)
	if registers.C_flag() {
		cc = 1
	}
	a = (a << 1) | cc
	registers.SetA(uint(a))

	c := false
	if r == 1 {
		c = true
	}

	registers.SetFlags(false, false, false, c)
}
func handleRRCA(instruction decoder.Instruction_t) {
	// 11111110 -> 11111101
	a := uint8(registers.A())
	r := a & 1
	a = (a >> 1) | (r << 7)
	registers.SetA(uint(a))

	c := false
	if r == 1 {
		c = true
	}

	registers.SetFlags(false, false, false, c)
}
func handleRRA(instruction decoder.Instruction_t) {
	// 11111110 -> 11111101
	a := uint8(registers.A())
	r := a & 1
	cc := uint8(0)
	if registers.C_flag() {
		cc = 1
	}
	a = (a >> 1) | (cc << 7)
	registers.SetA(uint(a))

	c := false
	if r == 1 {
		c = true
	}

	registers.SetFlags(false, false, false, c)
}

func handleRLC(instruction decoder.Instruction_t) {
	a := uint8(getValueByOperand(instruction.Operands[0]))
	r := (a >> 7) & 1
	a = (a << 1) | r

	setValueByOperand(instruction.Operands[0], uint(a))

	c := false
	if r == 1 {
		c = true
	}

	registers.SetFlags(a == 0, false, false, c)
}
func handleRL(instruction decoder.Instruction_t) {
	a := uint8(getValueByOperand(instruction.Operands[0]))
	r := (a >> 7) & 1
	cc := uint8(0)
	if registers.C_flag() {
		cc = 1
	}
	a = (a << 1) | cc

	setValueByOperand(instruction.Operands[0], uint(a))

	c := false
	if r == 1 {
		c = true
	}

	registers.SetFlags(a == 0, false, false, c)
}
func handleRRC(instruction decoder.Instruction_t) {
	// 11111110 -> 11111101
	a := uint8(getValueByOperand(instruction.Operands[0]))
	r := uint8(a) & 1
	a = (a >> 1) | (r << 7)

	setValueByOperand(instruction.Operands[0], uint(a))

	c := false
	if r == 1 {
		c = true
	}

	registers.SetFlags(a == 0, false, false, c)
}
func handleRR(instruction decoder.Instruction_t) {
	// 11111110 -> 11111101
	a := uint8(getValueByOperand(instruction.Operands[0]))
	r := a & 1
	cc := uint8(0)
	if registers.C_flag() {
		cc = 1
	}
	a = (a >> 1) | (cc << 7)

	setValueByOperand(instruction.Operands[0], uint(a))

	c := false
	if r == 1 {
		c = true
	}

	registers.SetFlags(a == 0, false, false, c)
}

func handleSLA(instruction decoder.Instruction_t) {
	a := uint8(getValueByOperand(instruction.Operands[0]))
	// set LSB to 0 -> 0xFE -> 11111110
	r := (a >> 7) & 1
	a = (uint8(a) << 1)

	setValueByOperand(instruction.Operands[0], uint(a))

	c := false
	if r == 1 {
		c = true
	}

	registers.SetFlags(a == 0, false, false, c)
}
func handleSRA(instruction decoder.Instruction_t) {
	a := uint8(getValueByOperand(instruction.Operands[0]))
	// MSB doesn't change
	r := uint8(a) & 1
	a = (uint8(a) >> 1) | (uint8(a) & 0x80)

	setValueByOperand(instruction.Operands[0], uint(a))

	c := false
	if r == 1 {
		c = true
	}

	registers.SetFlags(a == 0, false, false, c)
}
func handleSRL(instruction decoder.Instruction_t) {
	a := uint8(getValueByOperand(instruction.Operands[0]))
	// MSB set to 0
	r := uint8(a) & 1
	a = (uint8(a) >> 1) & 0x7F

	setValueByOperand(instruction.Operands[0], uint(a))

	c := false
	if r == 1 {
		c = true
	}

	registers.SetFlags(a == 0, false, false, c)
}

func handleBIT(instruction decoder.Instruction_t) {
	_b, err := (strconv.Atoi(instruction.Operands[0].Name))
	if err != nil || _b < 0 || _b > 7 {
		log.Fatal("Error during handling BIT")
	}

	value := uint8(getValueByOperand(instruction.Operands[1]))
	b := uint8(_b)

	registers.SetZeroFlag(utility.GetBit(uint(value), uint(b)) == 0)
	registers.SetSubstractionFlag(false)
	registers.SetHalfCarryFlag(true)
}
func handleSET(instruction decoder.Instruction_t) {
	_b, err := (strconv.Atoi(instruction.Operands[0].Name))
	if err != nil || _b < 0 || _b > 7 {
		log.Fatal("Error during handling BIT")
	}

	value := uint8(getValueByOperand(instruction.Operands[1]))
	b := uint8(_b)
	// SEt bit
	new_value := uint8(utility.SetBit(uint(value), uint(b)))

	setValueByOperand(instruction.Operands[1], uint(new_value))
}
func handleRES(instruction decoder.Instruction_t) {
	_b, err := (strconv.Atoi(instruction.Operands[0].Name))
	if err != nil || _b < 0 || _b > 7 {
		log.Fatal("Error during handling BIT")
	}

	value := uint8(getValueByOperand(instruction.Operands[1]))
	b := uint8(_b)
	// SEt bit
	new_value := uint8(utility.ClearBit(uint(value), uint(b)))

	setValueByOperand(instruction.Operands[1], uint(new_value))
}
