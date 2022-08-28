package registers

import (
	"fmt"
	"log"

	"github.com/giammirove/gbemu/internal/headers"
)

// the underscore before avoid exposing this constants
const _HALF_WORD = 8
const _HALF_WORD_MASK = 0xFF
const _MASK_LOW = 0x00FF
const _MASK_HI = 0xFF00

const _ZERO_FLAG = 7 // z
const _ZERO_FLAG_MASK = 128
const _SUB_FLAG = 6 // n
const _SUB_FLAG_MASK = 64
const _HALF_CARRY_FLAG = 5 // h
const _HALF_CARRY_FLAG_MASK = 32
const _CARRY_FLAG = 4 // c
const _CARRY_FLAG_MASK = 16

type registers_t struct {
	A  uint // accumulator & flags
	F  uint
	AF uint // 16 bit
	B  uint
	C  uint
	BC uint // 16 bit
	D  uint
	E  uint
	DE uint // 16 bit
	H  uint
	L  uint
	HL uint // 16 bit
	SP uint // stack pointer
	PC uint // program counter

}

var clock uint
var registers registers_t

func Init() {
	registers = registers_t{}
	SetPC(0x100)
	Reset()
	clock = 0
}

func Reset() {
	// DMG0
	if headers.IsGB() && !headers.IsCGB() {
		SetA(0x1)
		SetF(0x0)
		SetB(0xFF)
		SetC(0x13)
		SetD(0x0)
		SetE(0xC1)
		SetH(0x84)
		SetL(0x03)
		SetPC(0x0100)
		SetSP(0xFFFE)
	}
	if headers.IsCGB() {
		SetA(0x11)
		SetF(0x80)
		SetB(0x00)
		SetC(0x00)
		SetD(0xFF)
		SetE(0x08)
		SetH(0x00)
		SetL(0x0D)
		SetPC(0x0100)
		SetSP(0xFFFE)
	}

}

func IncrementClock(val ...uint) {
	if len(val) > 1 {
		log.Fatal("Too many arguments")
	}
	if len(val) == 1 {
		clock += val[0]
	} else {
		clock++
	}
}

func SetFlag(Set bool, bit int) {
	if Set {
		SetF(F() | 1<<bit)
	} else {
		SetF((F() | (1 << bit)) ^ (1 << bit))
	}
}

func SetFlags(z bool, n bool, h bool, c bool) {
	SetZeroFlag(z)
	SetSubstractionFlag(n)
	SetHalfCarryFlag(h)
	SetCarryFlag(c)
}

func SetZeroFlag(Set bool) {
	bit := _ZERO_FLAG
	SetFlag(Set, bit)
}
func Z_flag() bool { return F()&_ZERO_FLAG_MASK > 0 }

func SetSubstractionFlag(Set bool) {
	bit := _SUB_FLAG
	SetFlag(Set, bit)
}
func N_flag() bool { return F()&_SUB_FLAG_MASK > 0 }

func SetHalfCarryFlag(Set bool) {
	bit := _HALF_CARRY_FLAG
	SetFlag(Set, bit)
}
func H_flag() bool { return F()&_HALF_CARRY_FLAG_MASK > 0 }

func SetCarryFlag(Set bool) {
	bit := _CARRY_FLAG
	SetFlag(Set, bit)
}
func C_flag() bool { return F()&_CARRY_FLAG_MASK > 0 }

func SetA(val uint) {
	registers.A = val & _HALF_WORD_MASK
}

func SetF(val uint) {
	registers.F = val & _HALF_WORD_MASK
}

func SetAF(val uint) {
	registers.A = val & _MASK_HI >> _HALF_WORD
	registers.F = val & _MASK_LOW
}

func A() uint  { return registers.A & _HALF_WORD_MASK }
func F() uint  { return registers.F & _HALF_WORD_MASK }
func AF() uint { return registers.F | registers.A<<_HALF_WORD }

func SetB(val uint) {
	registers.B = val & _HALF_WORD_MASK
}

func SetC(val uint) {
	registers.C = val & _HALF_WORD_MASK
}

func SetBC(val uint) {
	registers.B = val & _MASK_HI >> _HALF_WORD
	registers.C = val & _MASK_LOW
}

func B() uint  { return registers.B & _HALF_WORD_MASK }
func C() uint  { return registers.C & _HALF_WORD_MASK }
func BC() uint { return registers.C | registers.B<<_HALF_WORD }

func SetD(val uint) {
	registers.D = val & _HALF_WORD_MASK
}

func SetE(val uint) {
	registers.E = val & _HALF_WORD_MASK
}

func SetDE(val uint) {
	registers.D = val & _MASK_HI >> _HALF_WORD
	registers.E = val & _MASK_LOW
}

func D() uint  { return registers.D & _HALF_WORD_MASK }
func E() uint  { return registers.E & _HALF_WORD_MASK }
func DE() uint { return registers.E | registers.D<<_HALF_WORD }

func SetH(val uint) {
	registers.H = val & _HALF_WORD_MASK
}

func SetL(val uint) {
	registers.L = val & _HALF_WORD_MASK
}

func SetHL(val uint) {
	registers.H = val & _MASK_HI >> _HALF_WORD
	registers.L = val & _MASK_LOW
}

func H() uint  { return registers.H & _HALF_WORD_MASK }
func L() uint  { return registers.L & _HALF_WORD_MASK }
func HL() uint { return registers.L | registers.H<<_HALF_WORD }

func SetSP(val uint) {
	registers.SP = val & 0xFFFF
}

func SetPC(val uint) {
	registers.PC = val & 0xFFFF
}

func SP() uint { return registers.SP & 0xFFFF }
func PC() uint { return registers.PC & 0xFFFF }

func DecrementSP() { registers.SP-- }
func DecrementPC() { registers.PC-- }
func IncrementSP() { registers.SP++ }
func IncrementPC() { registers.PC++ }

func Set(key string) func(val uint) {
	switch key {
	case "A":
		return SetA
	case "B":
		return SetB
	case "C":
		return SetC
	case "D":
		return SetD
	case "E":
		return SetE
	case "H":
		return SetH
	case "L":
		return SetL
	case "AF":
		return SetAF
	case "BC":
		return SetBC
	case "DE":
		return SetDE
	case "HL":
		return SetHL
	case "SP":
		return SetSP
	case "PC":
		return SetPC
	}
	log.Fatal("Register not found")
	return (func(val uint) {})
}
func Get(key string) func() uint {
	switch key {
	case "A":
		return A
	case "B":
		return B
	case "C":
		return C
	case "D":
		return D
	case "E":
		return E
	case "H":
		return H
	case "L":
		return L
	case "AF":
		return AF
	case "BC":
		return BC
	case "DE":
		return DE
	case "HL":
		return HL
	case "SP":
		return SP
	case "PC":
		return PC
	}
	log.Fatal("Register not found (", key, ")")
	return (func() uint { return 0 })
}

func GetNByte(key string) uint {
	switch key {
	case "A":
		return 1
	case "B":
		return 1
	case "C":
		return 1
	case "D":
		return 1
	case "E":
		return 1
	case "H":
		return 1
	case "L":
		return 1
	case "AF":
		return 2
	case "BC":
		return 2
	case "DE":
		return 2
	case "HL":
		return 2
	case "SP":
		return 2
	case "PC":
		return 2
	}
	log.Fatal("Register not found (", key, ")")
	return 0
}

func GetPair(key string) (func() uint, func() uint) {
	switch key {
	case "AF":
		return A, F
	case "BC":
		return B, C
	case "DE":
		return D, E
	case "HL":
		return H, L
	}
	log.Fatal("Register not found (", key, ")")
	return func() uint { return 0 }, func() uint { return 0 }
}
func SetPair(key string) (func(val uint), func(val uint)) {
	switch key {
	case "AF":
		return SetA, SetF
	case "BC":
		return SetB, SetC
	case "DE":
		return SetD, SetE
	case "HL":
		return SetH, SetL
	}
	log.Fatal("Register not found (", key, ")")
	return func(val uint) {}, func(val uint) {}
}

func GetFlag(key string) bool {
	switch key {
	case "NZ":
		return !Z_flag()
	case "Z":
		return Z_flag()
	case "NC":
		return !C_flag()
	case "C":
		return C_flag()
	}
	log.Fatal("Flag not found (", key, ")")
	return false
}

func Dump() {
	fmt.Printf("A: %02X F: %04b BC: %04X DE: %04X HL: %04X\n", A(), F()>>4, BC(), DE(), HL())
	// fmt.Printf("SP: %04X\n", SP())
}

func Test() {

	fmt.Println(A())
	SetCarryFlag(true)
	fmt.Println(C_flag())
	SetCarryFlag(false)
	fmt.Println(C_flag())
	SetBC(0x0F00)
	fmt.Printf("%016b\n", BC())
	fmt.Printf("%08b\n", B())
	fmt.Printf("%08b\n", C())
}
