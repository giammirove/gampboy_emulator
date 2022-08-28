package serial

import (
	"log"

	"github.com/giammirove/gbemu/internal/interrupts"
	"github.com/giammirove/gbemu/internal/utility"
)

const _SB = 0xFF01
const _SC = 0xFF02
const _START_ADDR = _SB
const _END_ADDR = _SC

var registers []uint

// TOOD: add external clock
func Init() {
	registers = make([]uint, _END_ADDR-_START_ADDR+1)
	WriteToMemory(_SB, 0x00)
	WriteToMemory(_SC, 0x7E)
}

func IsSerialAddr(addr uint) bool {
	return addr >= _START_ADDR && addr <= _END_ADDR
}

func ReadFromMemory(addr uint) uint {
	if addr < _START_ADDR || addr > _END_ADDR {
		log.Fatalf("Invalid sound address (0x%8X)", addr)
	}
	return registers[addr-_START_ADDR]
}
func WriteToMemory(addr uint, value uint) {
	if addr < _START_ADDR || addr > _END_ADDR {
		log.Fatalf("Invalid sound address (0x%8X)", addr)
	}
	registers[addr-_START_ADDR] = value
	if addr == _SC && isTransferring() {
		resetSC()
		interrupts.RequestInterruptSerial()
	}
}

func resetSC() {
	registers[_SC-_START_ADDR] = 0x81
}
func isTransferring() bool {
	return utility.TestBit(ReadFromMemory(_SB), 7)
}
