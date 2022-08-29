package cpu

import (
	"github.com/giammirove/gampboy_emulator/internal/decoder"
	"github.com/giammirove/gampboy_emulator/internal/registers"
	"github.com/giammirove/gampboy_emulator/internal/utility"
)

func handleINC(instruction decoder.Instruction_t) {

	switch instruction.Opcode {
	case 0x3C, 0x04, 0x0C, 0x14, 0x1C, 0x24, 0x2C, 0x34:
		handleINC_n(instruction)
		break
	case 0x03, 0x13, 0x23, 0x33:
		handleINC_nn(instruction)
		break
	default:
		notHandled(instruction)
	}
}
func handleINC_n(instruction decoder.Instruction_t) {
	value := uint8(getValueByOperand(instruction.Operands[0]))

	registers.SetSubstractionFlag(false)
	registers.SetHalfCarryFlag(utility.CheckHalfCarry(uint(value), 1))

	value++
	setValueByOperand(instruction.Operands[0], uint(value))

	registers.SetZeroFlag(uint8(value) == 0)
}
func handleINC_nn(instruction decoder.Instruction_t) {
	Cycle(4)
	value := uint16(getValueByOperand(instruction.Operands[0]))
	value++
	setValueByOperand(instruction.Operands[0], uint(value))
}

func handleDEC(instruction decoder.Instruction_t) {

	switch instruction.Opcode {
	case 0x3D, 0x05, 0x0D, 0x15, 0x1D, 0x25, 0x2D, 0x35:
		handleDEC_n(instruction)
		break
	case 0x0B, 0x1B, 0x2B, 0x3B:
		handleDEC_nn(instruction)
		break
	default:
		notHandled(instruction)
	}
}
func handleDEC_n(instruction decoder.Instruction_t) {
	value := uint16(getValueByOperand(instruction.Operands[0]) & 0xFF)

	registers.SetSubstractionFlag(true)
	registers.SetHalfCarryFlag(utility.CheckHalfCarrySub(uint(value), 1))

	value--
	setValueByOperand(instruction.Operands[0], uint(value))

	registers.SetZeroFlag(uint8(value) == 0)
}
func handleDEC_nn(instruction decoder.Instruction_t) {
	Cycle(4)
	value := uint16(getValueByOperand(instruction.Operands[0])) - 1
	setValueByOperand(instruction.Operands[0], uint(value))
}

func handleCP(instruction decoder.Instruction_t) {
	value := int16(getValueByOperand(instruction.Operands[0])) & 0xFF
	a := int16(registers.A())
	n := a - value
	registers.SetZeroFlag(n == 0)
	registers.SetSubstractionFlag(true)
	registers.SetHalfCarryFlag(utility.CheckHalfCarrySub(uint(a), uint(value)))
	registers.SetCarryFlag(n < 0)
}

func handleOR(instruction decoder.Instruction_t) {
	value := getValueByOperand(instruction.Operands[0])
	res := uint8(registers.A()) | uint8(value)

	registers.SetZeroFlag(uint8(res) == 0)
	registers.SetSubstractionFlag(false)
	registers.SetHalfCarryFlag(false)
	registers.SetCarryFlag(false)

	registers.SetA(uint(res))
}
func handleXOR(instruction decoder.Instruction_t) {
	value := getValueByOperand(instruction.Operands[0])
	res := uint8(registers.A()) ^ uint8(value)

	registers.SetZeroFlag(uint8(res) == 0)
	registers.SetSubstractionFlag(false)
	registers.SetHalfCarryFlag(false)
	registers.SetCarryFlag(false)

	registers.SetA(uint(res))
}
func handleAND(instruction decoder.Instruction_t) {
	value := getValueByOperand(instruction.Operands[0])
	res := uint8(registers.A()) & uint8(value)

	registers.SetZeroFlag(uint8(res) == 0)
	registers.SetSubstractionFlag(false)
	registers.SetHalfCarryFlag(true)
	registers.SetCarryFlag(false)

	registers.SetA(uint(res))
}

func handleADD_A_n(instruction decoder.Instruction_t) {
	value := getValueByOperand(instruction.Operands[1])

	registers.SetSubstractionFlag(false)
	registers.SetHalfCarryFlag(utility.CheckHalfCarry(registers.A(), value))
	registers.SetCarryFlag(utility.CheckCarry(registers.A(), value))

	registers.SetA(registers.A() + value)

	registers.SetZeroFlag(registers.A() == 0)
}
func handleADC_A_n(instruction decoder.Instruction_t) {
	value := uint16(getValueByOperand(instruction.Operands[1]))
	c := uint(0)
	if registers.C_flag() {
		c = 1
	}

	registers.SetSubstractionFlag(false)
	registers.SetHalfCarryFlag(utility.CheckTriHalfCarry(registers.A(), uint(value), c))
	registers.SetCarryFlag(utility.CheckTriCarry(registers.A(), uint(value), c))

	registers.SetA(registers.A() + uint(value) + c)

	registers.SetZeroFlag(registers.A() == 0)
}
func handleADD_HL_n(instruction decoder.Instruction_t) {
	Cycle(4)
	value := getValueByOperand(instruction.Operands[1])

	registers.SetSubstractionFlag(false)
	registers.SetHalfCarryFlag(utility.CheckHalfCarry16bit(registers.HL(), value))
	registers.SetCarryFlag(utility.CheckCarry16bit(registers.HL(), value))

	registers.SetHL(registers.HL() + value)
}
func handleADD_SP_n(instruction decoder.Instruction_t) {
	Cycle(8)
	// value is signed
	value := int8(getValueByOperand(instruction.Operands[1]))

	registers.SetZeroFlag(false)
	registers.SetSubstractionFlag(false)
	registers.SetHalfCarryFlag(utility.CheckHalfCarry(registers.SP(), uint(value)))
	registers.SetCarryFlag(utility.CheckCarry(registers.SP(), uint(value)))

	registers.SetSP(uint(int16(registers.SP()) + int16(value)))
}

func handleSUB_A_n(instruction decoder.Instruction_t) {
	value := getValueByOperand(instruction.Operands[0])

	registers.SetSubstractionFlag(true)
	registers.SetHalfCarryFlag(utility.CheckHalfCarrySub(registers.A(), value))
	registers.SetCarryFlag(utility.CheckCarrySub(registers.A(), value))

	registers.SetA(registers.A() - value)

	registers.SetZeroFlag(registers.A() == 0)
}
func handleSBC_A_n(instruction decoder.Instruction_t) {
	value := uint16(getValueByOperand(instruction.Operands[1]))
	c := uint(0)
	if registers.C_flag() {
		c = 1
	}

	registers.SetSubstractionFlag(true)
	registers.SetHalfCarryFlag(utility.CheckTriHalfCarrySub(registers.A(), uint(value), c))
	registers.SetCarryFlag(utility.CheckTriCarrySub(registers.A(), uint(value), c))

	registers.SetA(registers.A() - uint(value) - c)

	registers.SetZeroFlag(registers.A() == 0)
}

func handleADD(instruction decoder.Instruction_t) {
	switch instruction.Opcode {
	case 0x87, 0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0xC6:
		handleADD_A_n(instruction)
		break
	case 0x8F, 0x88, 0x89, 0x8A, 0x8B, 0x8C, 0x8D, 0x8E, 0xCE:
		handleADC_A_n(instruction)
		break
	case 0x09, 0x19, 0x29, 0x39:
		handleADD_HL_n(instruction)
		break
	case 0xE8:
		handleADD_SP_n(instruction)
		break
	default:
		notHandled(instruction)
	}
}

func handleSUB(instruction decoder.Instruction_t) {
	switch instruction.Opcode {
	case 0x97, 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0xD6:
		handleSUB_A_n(instruction)
		break
	case 0x9F, 0x98, 0x99, 0x9A, 0x9B, 0x9C, 0x9D, 0x9E, 0xDE:
		handleSBC_A_n(instruction)
		break
	default:
		notHandled(instruction)
	}
}
