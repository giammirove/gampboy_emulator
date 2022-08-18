package sound

import "log"

var registers []uint

const _START_ADDR = 0xFF10
const _END_ADDR = 0xFF26

func Init() {
	registers = make([]uint, _END_ADDR-_START_ADDR+1)
}

func IsSoundAddr(addr uint) bool {
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
