package sound

import "log"

var registers []uint

const _START_ADDR = 0xFF10
const _END_ADDR = 0xFF26

func Init() {
	registers = make([]uint, _END_ADDR-_START_ADDR+1)
	WriteToMemory(0xFF10, 0x80)
	WriteToMemory(0xFF11, 0xBF)
	WriteToMemory(0xFF12, 0xF3)
	WriteToMemory(0xFF13, 0xFF)
	WriteToMemory(0xFF14, 0xBF)
	WriteToMemory(0xFF15, 0xFF)
	WriteToMemory(0xFF16, 0x3F)
	WriteToMemory(0xFF17, 0x00)
	WriteToMemory(0xFF18, 0xFF)
	WriteToMemory(0xFF19, 0xBF)
	WriteToMemory(0xFF1A, 0x7F)
	WriteToMemory(0xFF1B, 0xFF)
	WriteToMemory(0xFF1C, 0x9F)
	WriteToMemory(0xFF1D, 0xFF)
	WriteToMemory(0xFF1E, 0xBF)
	WriteToMemory(0xFF1F, 0xFF)
	WriteToMemory(0xFF20, 0xFF)
	WriteToMemory(0xFF21, 0x00)
	WriteToMemory(0xFF22, 0x00)
	WriteToMemory(0xFF23, 0xBF)
	WriteToMemory(0xFF24, 0x77)
	WriteToMemory(0xFF25, 0xF3)
	WriteToMemory(0xFF26, 0xF1)
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
