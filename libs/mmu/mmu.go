package mmu

import (
	"log"

	"github.com/giammirove/gbemu/libs/interrupts"
	"github.com/giammirove/gbemu/libs/joypad"
	"github.com/giammirove/gbemu/libs/ppu"
	"github.com/giammirove/gbemu/libs/serial"
	"github.com/giammirove/gbemu/libs/sound"
	"github.com/giammirove/gbemu/libs/timer"
	"github.com/giammirove/gbemu/libs/utility"
)

type memory_region struct {
	start uint
	end   uint
}

const _BYTE_SIZE = 8

const _ROM0_START = 0x0000
const _ROM0_END = 0x3FFF
const _ROM1_START = 0x4000
const _ROM1_END = 0x7FFF
const _VRAM_START = 0x8000
const _VRAM_END = 0x9FFF
const _ERAM_START = 0xA000
const _ERAM_END = 0xBFFF
const _RAM_START = 0xC000
const _RAM_END = 0xDFFF
const _ECHORAM_START = 0xE000
const _ECHORAM_END = 0xFDFF
const _OAM_START = 0xFE00
const _OAM_END = 0xFE9F
const _HRAM_START = 0xFF80
const _HRAM_END = 0xFFFE

var ROM []byte
var HRAM [_HRAM_END - _HRAM_START + 1]byte

var rom_path string

func InitMMU(rom []byte, path string) {
	ROM = rom
	rom_path = path

	InitMBC()
}

func readFromHighRAM(addr uint) byte {
	if addr < _HRAM_START || addr > _HRAM_END {
		log.Fatal("Address not in HRAM boundary (Read)")
	}
	return HRAM[addr-_HRAM_START]
}
func writeToHighRAM(addr uint, value byte) {
	if addr < _HRAM_START || addr > _HRAM_END {
		log.Fatal("Address not in HRAM boundary (Write)")
	}
	HRAM[addr-_HRAM_START] = value
}

func readByteMemory(addr uint) byte {
	switch addr & 0xF000 {
	case 0x0000:
		fallthrough
	case 0x1000:
		fallthrough
	case 0x2000:
		fallthrough
	case 0x3000:
		return ReadFromRomMemory(addr)
	case 0x4000:
		fallthrough
	case 0x5000:
		fallthrough
	case 0x6000:
		fallthrough
	case 0x7000:
		return ReadFromRomMemory(addr)
	case 0x8000:
		fallthrough
	case 0x9000:
		return ppu.ReadFromVRAMMemory(addr)
	case 0xA000:
		fallthrough
	case 0xB000:
		return ReadFromRamMemory(addr)
	case 0xC000:
		fallthrough
	case 0xD000:
		return ReadFromRamMemory(addr)
	}

	if interrupts.IsInterruptAddr(addr) {
		return byte(interrupts.ReadFromMemory(addr))
	} else if addr <= _ECHORAM_END {
		// return 0xFF
		// return readFromEchoRAM(addr)
		return ReadFromRamMemory(addr - (_ECHORAM_START - _RAM_START))
	} else if addr <= _OAM_END {
		// TODO: re-add
		// if ppu.GetDMATransferring() {
		// 	return 0xFF
		// }
		return ppu.ReadFromOAMMemory(addr)
	} else if addr <= 0xFEFF {
		log.Printf("Use of this area is prohibited %04X\n", addr)
		return 0
	} else if addr <= 0xFF7F {
		if joypad.IsJoypadAddr(addr) {
			return byte(joypad.ReadFromMemory(addr))
		}
		if serial.IsSerialAddr(addr) {
			return byte(serial.ReadFromMemory(addr))
		}
		if interrupts.IsInterruptAddr(addr) {
			return byte(interrupts.ReadFromMemory(addr))
		}
		if timer.IsTimerAddr(addr) {
			return byte(timer.ReadFromMemory(addr))
		}
		if ppu.IsLCDAddr(addr) {
			return byte(ppu.ReadFromLCDMemory(addr))
		}
		if sound.IsSoundAddr(addr) {
			return byte(sound.ReadFromMemory(addr))
		}
		// log.Printf("I/O not handled (Read) (0x%02X)", addr)
		return 0
	} else if addr <= _HRAM_END {
		return readFromHighRAM(addr)
	}

	log.Fatalf("Invalid address (Read) (0x%08X)", addr)
	return 0
}

func readWordMemory(addr uint) uint16 {
	return uint16(readByteMemory(addr)) + uint16(readByteMemory(addr+1))<<_BYTE_SIZE
}

func ReadFromMemory(addr uint, bytes ...uint) uint {
	if len(bytes) > 1 {
		log.Fatal("Too many arguments")
	}

	if len(bytes) == 1 {
		if bytes[0] > 2 {
			log.Fatal("Too many bytes to read")
		}
		if bytes[0] == 1 {
			return uint(readByteMemory(addr))
		} else {
			return uint(readWordMemory(addr))
		}
	}

	return uint(readByteMemory(addr))
}

func writeByteMemory(addr uint, value byte) {
	switch addr & 0xF000 {
	case 0x0:
		fallthrough
	case 0x1000:
		fallthrough
	case 0x2000:
		fallthrough
	case 0x3000:
		WriteToRomMemory(addr, value)
		return
	case 0x4000:
		fallthrough
	case 0x5000:
		fallthrough
	case 0x6000:
		fallthrough
	case 0x7000:
		WriteToRomMemory(addr, value)
		return
	case 0x8000:
		fallthrough
	case 0x9000:
		ppu.WriteToVRAMMemory(addr, value)
		return
	case 0xA000:
		fallthrough
	case 0xB000:
		WriteToRamMemory(addr, value)
		return
	case 0xC000:
		fallthrough
	case 0xD000:
		WriteToRamMemory(addr, value)
		return
	}

	if interrupts.IsInterruptAddr(addr) {
		interrupts.WriteToMemory(addr, uint(value))
		return
	} else if addr <= _ECHORAM_END {
		// writeToEchoRAM(addr, value)
		WriteToRamMemory(addr-(_ECHORAM_START-_RAM_START), value)
		return
	} else if addr <= _OAM_END {
		// TODO: re-add
		// if ppu.GetDMATransferring() {
		// 	return
		// }
		ppu.WriteToOAMMemory(addr, value)
		return
	} else if addr <= 0xFEFF {
		log.Printf("Use of this area is prohibited %04X\n", addr)
		return
	} else if addr <= 0xFF7F {
		if joypad.IsJoypadAddr(addr) {
			joypad.WriteToMemory(addr, uint(value))
			return
		}
		if serial.IsSerialAddr(addr) {
			serial.WriteToMemory(addr, uint(value))
			return
		}
		if timer.IsTimerAddr(addr) {
			timer.WriteToMemory(addr, uint(value))
			return
		}
		if ppu.IsLCDAddr(addr) {
			ppu.WriteToLCDMemory(addr, uint(value))
			return
		}
		if sound.IsSoundAddr(addr) {
			sound.WriteToMemory(addr, uint(value))
			return
		}
		// log.Printf("I/O not handled (0x%02X)", addr)
		return
	} else if addr <= _HRAM_END {
		writeToHighRAM(addr, value)
		return
	} else if interrupts.IsInterruptAddr(addr) {
		interrupts.WriteToMemory(addr, uint(value))
		return
	}

	log.Fatalf("Invalid address (Write) (0x%08X)", addr)
}
func writeWordMemory(addr uint, bytes []byte) {
	writeByteMemory(addr, bytes[0])
	writeByteMemory(addr+1, bytes[1])
}

func WriteToMemory(addr uint, value uint, bytes ...uint) {
	if len(bytes) > 1 {
		log.Fatal("Too many arguments")
	}

	if len(bytes) == 1 {
		if bytes[0] > 2 {
			log.Fatal("Too many bytes (Write)")
		}
		if bytes[0] == 2 {
			hi, low := utility.GetHiLow(uint16(value))
			writeWordMemory(addr, []byte{low, hi})
		} else {
			writeByteMemory(addr, byte(value))
		}
	} else {
		writeByteMemory(addr, byte(value))
	}

}
