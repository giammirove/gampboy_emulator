package timer

import (
	"log"

	"github.com/giammirove/gbemu/internal/headers"
	"github.com/giammirove/gbemu/internal/interrupts"
	"github.com/giammirove/gbemu/internal/utility"
)

var registers []uint

const _START_ADDR = 0xFF04
const _END_ADDR = 0xFF07

const _DIV = 0xFF04
const _TIMA = 0xFF05
const _TMA = 0xFF06
const _TAC = 0xFF07

const _DIV_I = 0
const _TIMA_I = 1
const _TMA_I = 2
const _TAC_I = 3

const _RESET_TIME_CLOCK = 4

var div_internal = uint(0)
var tima_clock = 0

var resetting_tima = false
var resetting_tima_ticks = 0
var old_state = false

// timer enable bit of TAC
const _TAC_TE_BIT = 2

// TAC = CPU clock / _TAC_IC_SELECT[i]
var _TAC_IC_SELECT = [4]int{1024, 16, 64, 256}

func Init() {
	registers = make([]uint, _END_ADDR-_START_ADDR+1)
	if headers.IsGB() {
		registers[_DIV_I] = 0x18
		registers[_TAC_I] = 0xF8
	}
	resetting_tima = false
	resetting_tima_ticks = 0
	old_state = true
}

func Tick() {

	div_internal++
	// increment div every 64 m-cycles
	// TODO: why with 63 works???
	if div_internal%(63*4) == 0 {
		registers[_DIV_I] = (registers[_DIV_I] + 1) & 0xFFFF
	}
	// TODO: handle different cpu speed
	freq := GetClockFreq()

	if resetting_tima {
		if resetting_tima_ticks == _RESET_TIME_CLOCK {
			interrupts.RequestInterruptTimer()
		}
		if resetting_tima_ticks == _RESET_TIME_CLOCK+1 {
			registers[_TIMA_I] = registers[_TMA_I]
		}
		if resetting_tima_ticks == _RESET_TIME_CLOCK+2 {
			resetting_tima = false
			registers[_TIMA_I] = registers[_TMA_I]
			resetting_tima_ticks = 0
		}
		resetting_tima_ticks++
	} else if div_internal%freq == 0 && GetTACEnable() {
		stepTIMA()
		registers[_DIV_I] = 0
		div_internal = 0
	}

}

func IsTimerAddr(addr uint) bool {
	return addr >= _START_ADDR && addr <= _END_ADDR
}

func ReadFromMemory(addr uint) uint {
	if addr < _START_ADDR || addr > _END_ADDR {
		log.Fatalf("Invalid timer address (0x%8X)", addr)
	}
	ret := registers[addr-_START_ADDR]
	// reset every time
	if addr == _DIV {
		// ret = ret >> 8
	}
	if addr == _TAC {
		ret = ret | 0b11111000
	}
	return ret
}
func WriteToMemory(addr uint, value uint) {
	if addr < _START_ADDR || addr > _END_ADDR {
		log.Fatalf("Invalid timer address (0x%8X)", addr)
	}
	// reset every time
	if addr == _DIV {
		// obscure behaviour
		if registers[_DIV_I]&0b100000000 == 0b100000000 {
			stepTIMA()
		}
		value = 0
		div_internal = 0
	}
	registers[addr-_START_ADDR] = value
}

var Cycle func(val uint)

func stepTIMA() {
	// not really necessary but lower numbers are easier to manage
	// registers[_DIV_I] = 0
	registers[_TIMA_I]++
	if registers[_TIMA_I] >= 0xFF {
		resetTIMA()
	}
}

func resetTIMA() {
	// reset time takes 1 cycle (4 clocks)
	resetting_tima = true
}
func ResetDIV() {
	div_internal = 0
	WriteToMemory(_DIV, 0)
}
func GetDIV() uint {
	return ReadFromMemory(_DIV)
}
func GetDIVInternal() uint {
	return div_internal
}

func GetTACEnable() bool {
	return utility.GetBit(registers[_TAC_I], _TAC_TE_BIT) == 0x1
}

func GetTACSelect() uint {
	// only 0-1 bits
	return registers[_TAC_I] & 0x3
}

func GetClockFreq() uint {
	return uint(_TAC_IC_SELECT[GetTACSelect()])
}
