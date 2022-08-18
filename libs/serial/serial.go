package serial

import (
	"log"
)

var registers []uint

const _START_ADDR = 0xFF01
const _END_ADDR = 0xFF02

func Init() {
	registers = make([]uint, _END_ADDR-_START_ADDR+1)
	WriteToMemory(0xFF01, 0x00)
	WriteToMemory(0xFF02, 0x7E)
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
}
