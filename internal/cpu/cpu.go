package cpu

import (
	"fmt"
	"log"
	"strings"

	decoder "github.com/giammirove/gampboy_emulator/internal/decoder"
	"github.com/giammirove/gampboy_emulator/internal/interrupts"
	"github.com/giammirove/gampboy_emulator/internal/mmu"
	"github.com/giammirove/gampboy_emulator/internal/ppu"
	registers "github.com/giammirove/gampboy_emulator/internal/registers"
	"github.com/giammirove/gampboy_emulator/internal/timer"
	"github.com/giammirove/gampboy_emulator/internal/utility"
)

const _LD = "LD"
const _LDH = "LDH"
const _PUSH = "PUSH"
const _POP = "POP"
const _JP = "JP"
const _JR = "JR"
const _CALL = "CALL"
const _RST = "RST"
const _RET = "RET"
const _RETI = "RETI"

// ALU
const _AND = "AND"
const _OR = "OR"
const _XOR = "XOR"
const _INC = "INC"
const _DEC = "DEC"
const _CP = "CP"
const _ADD = "ADD"
const _ADC = "ADC"
const _SUB = "SUB"
const _SBC = "SBC"

// MISC
const _SWAP = "SWAP"
const _DAA = "DAA"
const _CPL = "CPL"
const _CCF = "CCF"
const _SCF = "SCF"
const _NOP = "NOP"
const _HALT = "HALT"
const _STOP = "STOP"
const _DI = "DI"
const _EI = "EI"

//  ROTATE & SHIFTS
const _RLCA = "RLCA"
const _RLA = "RLA"
const _RRCA = "RRCA"
const _RRA = "RRA"
const _RLC = "RLC"
const _RL = "RL"
const _RRC = "RRC"
const _RR = "RR"
const _SLA = "SLA"
const _SRA = "SRA"
const _SRL = "SRL"

// BIT
const _BIT = "BIT"
const _SET = "SET"
const _RES = "RES"

const CLOCK_SPEED = 4194304
const MAX_FPS = 120
const CPS = CLOCK_SPEED / MAX_FPS

var msg [1024]byte
var msg_size = 0

var halted bool

func InitCPU() {
	registers.Init()
	msg = [1024]byte{0}
	halted = false
}
func SetHalted(val bool) {
	halted = val
}
func GetHalted() bool {
	return halted
}

var ticks int = 0

var DEBUG = false
var PAUSE = false
var MANUAL = false

func ToggleDebugMode() {
	if DEBUG {
		DEBUG = false
	} else {
		DEBUG = true
	}
}
func TogglePauseMode() {
	if PAUSE {
		PAUSE = false
	} else {
		PAUSE = true
		log.Printf("ticks %d\n", ticks)
	}
}
func ToggleManualMode() {
	if MANUAL {
		MANUAL = false
	} else {
		MANUAL = true
	}
}

func Run() {
	// ticker := time.NewTicker(time.Second / time.Duration(MAX_FPS))
	// for range ticker.C {
	// 	for i := 0; i < CPS; i++ {
	for {
		if !PAUSE {

			if !GetHalted() {
				if ppu.IsGDMATransferring() {
					Cycle(4)
				} else {
					// dbgUpdate()
					// dbgPrint()
					if interrupts.GetIF()&interrupts.GetIE() != 0 && !interrupts.GetIME() {
						// utility.WaitHere("halt bug")
					}

					// pre_d := ppu.GetLY()

					addr := registers.PC()
					saved := addr
					instruction := decoder.Decode(&addr)

					registers.SetPC(addr)

					// if ticks == 0x315BA {
					// 	instruction.Operands[1].Value += 6
					// }
					if DEBUG {
						fmt.Printf("%05X - $%05X: ", ticks, saved)
						fmt.Printf("%-19s (%02X %02X) ", decoder.PrintInstrunction(instruction), mmu.ReadFromMemoryCPU(saved+1), mmu.ReadFromMemoryCPU(saved+2))
						registers.Dump()
						fmt.Printf("%-43s SP: %04X PC: %04X ROM: %02d RAM: %02d\n", "", registers.SP(), registers.PC(), mmu.GetRomBank(), mmu.GetRamBank())
					}
					execute(instruction)
					if MANUAL {
						utility.WaitHere()
					}
					if instruction.Mnemonic == "LD" && len(instruction.Operands) > 1 && instruction.Operands[0].Name == "B" && instruction.Operands[1].Value == 0x00 && instruction.Operands[1].Immediate {
						// utility.WaitHere()
					}
					// if len(instruction.Operands) > 0 && instruction.Operands[0].Name == "B" {
					// 	utility.WaitHere()
					// }
					// if ticks >= 0x2DD58 && pre_d != ppu.GetLY() {
					// 	fmt.Printf("ly %d -> %d\n", pre_d, ppu.GetLY())
					// 	// utility.WaitHere()
					// }

					if ticks == 245000 {
						// utility.WaitHere()
					}

					ticks++

				}

			} else {
				Cycle(4)
				if interrupts.GetIF()&interrupts.GetIE()&0b11111 != 0 {
					SetHalted(false)
				}
			}
			if interrupts.GetIME() {
				if interrupts.HandleInterrupts() {
					SetHalted(false)
				}
			}
			// EI  is delayed by one instruction
			// But if EI is followed immediately by DI does not allow any interrupts
			if interrupts.GetPendingIME() {
				interrupts.SetIME(1)
				interrupts.SetPendingIME(0)
			}

			if GetHalted() && (interrupts.GetIE()|interrupts.GetIF())&0b11111 == 0x0 {
				log.Fatal("HALT")
			}
		}
	}
	// 	}
	// }
}

// val -> T-Cycle = M-Cycle * 4
// M-Cycle = cpu Cycle
func Cycle(val uint) {
	m_cycle := int(val / 4)
	// M-Cycle
	for i := 0; i < m_cycle; i++ {
		// T-Cycle
		for j := 0; j < 4; j++ {
			timer.Tick()
			if ppu.IsDoubleSpeed() {
				timer.Tick()
			}
			ppu.LCDTick()
		}
		ppu.DMATick()
		if ppu.IsDoubleSpeed() {
			ppu.DMATick()
		}
		//TODO:  faster
		ppu.GDMATransfer()
	}
}

func dbgUpdate() {
	if mmu.ReadFromMemoryCPU(0xFF02) == 0x81 {
		// log.Fatal("DBG UPDATE")
		c := mmu.ReadFromMemoryCPU(0xFF01)
		msg[msg_size] = byte(c)
		msg_size++
		mmu.WriteToMemoryCPU(0xFF02, 0)
	}
}

func dbgPrint() {
	if msg[0] != 0 {
		fmt.Printf("DBG : %s\n", msg)
		m := string(msg[:])
		if strings.Contains(m, "Failed") || strings.Contains(m, "Passed") {
			// fmt.Printf("%s\n", msg)
			// log.Fatal("TEST COMPLETE")
		}
		// log.Fatal("DGB RECEIVED")
	}
}

func debug(instruction decoder.Instruction_t) {
	if instruction.Opcode == 0xDE && false {
		fmt.Println(instruction)
		fmt.Printf("A: %02X\n", registers.A())
		// fmt.Printf("SP: %04X HL: %04X\n", registers.SP(), registers.HL())
		utility.WaitHere()
	}
}

func execute(instruction decoder.Instruction_t) {

	switch instruction.Mnemonic {
	case _LD, _LDH:
		handleLD(instruction)
		break
	case _PUSH:
		handlePUSH(instruction)
		break
	case _POP:
		handlePOP(instruction)
		break
	case _JP, _JR:
		handleJP(instruction)
		break
	case _CALL:
		handleCALL(instruction)
		break
	case _RST:
		handleRST(instruction)
		break
	case _RET:
		handleRET(instruction)
		break
	case _RETI:
		handleRETI(instruction)
		break
	case _AND:
		handleAND(instruction)
		break
	case _OR:
		handleOR(instruction)
		break
	case _XOR:
		handleXOR(instruction)
		break
	case _INC:
		handleINC(instruction)
		break
	case _DEC:
		handleDEC(instruction)
		break
	case _CP:
		handleCP(instruction)
		break
	case _ADD, _ADC:
		handleADD(instruction)
		break
	case _SUB, _SBC:
		handleSUB(instruction)
		break

	case _SWAP:
		handleSWAP(instruction)
		break
	case _DAA:
		handleDAA(instruction)
		break
	case _CPL:
		handleCPL(instruction)
		break
	case _CCF:
		handleCCF(instruction)
		break
	case _SCF:
		handleSCF(instruction)
		break
	case _NOP:
		handleNOP(instruction)
		break
	case _HALT:
		handleHALT(instruction)
		break
	case _STOP:
		handleSTOP(instruction)
		break
	case _DI:
		handleDI(instruction)
		break
	case _EI:
		handleEI(instruction)
		break

	case _RLCA:
		handleRLCA(instruction)
		break
	case _RLA:
		handleRLA(instruction)
		break
	case _RRCA:
		handleRRCA(instruction)
		break
	case _RRA:
		handleRRA(instruction)
		break
	case _RLC:
		handleRLC(instruction)
		break
	case _RL:
		handleRL(instruction)
		break
	case _RRC:
		handleRRC(instruction)
		break
	case _RR:
		handleRR(instruction)
		break
	case _SLA:
		handleSLA(instruction)
		break
	case _SRA:
		handleSRA(instruction)
		break
	case _SRL:
		handleSRL(instruction)
		break

	case _BIT:
		handleBIT(instruction)
		break
	case _SET:
		handleSET(instruction)
		break
	case _RES:
		handleRES(instruction)
		break
	default:
		log.Fatal("Not handled (", instruction.Mnemonic, ")")
	}

}

func PrintStack() {
	for i := 0; i < 10; i++ {
		fmt.Printf("%02X ", mmu.ReadFromMemoryCPU(registers.SP()-uint(i)))
	}
	fmt.Printf("\n")
}

func StackPush(value uint, bytes ...uint) {
	if len(bytes) > 1 {
		log.Fatal("Too many arguments")
	}
	if len(bytes) == 1 {
		b := bytes[0]
		if b > 2 || b <= 0 {
			log.Fatal("Wrong number of bytes to push")
		}
		if b == 1 {
			registers.DecrementSP()
			Cycle(4)
			mmu.WriteToMemoryCPU(registers.SP(), uint(uint8(value)))
		} else {
			hi, low := utility.GetHiLow(uint16(value))
			Cycle(4)
			registers.DecrementSP()
			mmu.WriteToMemoryCPU(registers.SP(), uint(hi))
			Cycle(4)
			registers.DecrementSP()
			mmu.WriteToMemoryCPU(registers.SP(), uint(low))
		}
	} else {
		Cycle(4)
		registers.DecrementSP()
		mmu.WriteToMemoryCPU(registers.SP(), uint(uint8(value)))
	}
}

func StackPOPSingle() uint {
	Cycle(4)
	value := mmu.ReadFromMemoryCPU(registers.SP())
	registers.IncrementSP()
	return value
}
func StackPOP() uint {
	low := StackPOPSingle()
	hi := StackPOPSingle()
	return uint(utility.SetHiLow(uint8(hi), uint8(low)))
}
