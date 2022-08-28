package mmu

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/giammirove/gbemu/internal/headers"
	"github.com/giammirove/gbemu/internal/ppu"
)

// same for MBC1 , MBC3 (also for timer)
const _RAM_ENABLE_END = 0x1FFF
const _RAM_ENABLE_VALUE = 0xA
const _RAM_DISABLE_VALUE = 0x0

// same for MBC1 , MBC3
const _ROM_BANK_NUMBER_START = 0x2000
const _ROM_BANK_NUMBER_END = 0x3FFF

// same for MBC1 , MBC3
const _RAM_BANK_NUMBER_START = 0x4000
const _RAM_BANK_NUMBER_END = 0x5FFF

const _BANKING_MODE_START = 0x6000
const _BANKING_MODE_END = 0x7FFF

const _ROM_BANK_SIZE = 16 * 1024
const _RAM_BANK_SIZE = 8 * 1024

const _RTC_REGISTERS_NUM = 5

// manage rom banks and external ram
var WRAM [_RAM_END - _RAM_START + 1]byte
var WRAM_CGB [8][_RAM_CGB_END - _RAM_CGB_START + 1]byte
var ERAM [_ERAM_END - _ERAM_START + 1]byte

// 8 KiB each
var ERAM_BANKS []byte

// false = default, true = advanced banking mode
var banking_mode bool
var ram_enabled bool
var rom_bank uint
var ram_bank uint

var rtc uint8
var rtc_active bool
var rtc_registers [_RTC_REGISTERS_NUM]uint8
var rtc_latched bool
var rtc_registers_latched [_RTC_REGISTERS_NUM]uint8

var rom_memory_path string
var save_needed bool

func InitMBC() {

	// ERAM_BANKS = make([]byte, headers.GetRamBankNumber())

	banking_mode = false
	ram_enabled = false
	rom_bank = 1
	ram_bank = 0

	rtc = 0
	rtc_active = false
	rtc_registers = [_RTC_REGISTERS_NUM]uint8{}
	rtc_latched = false
	rtc_registers_latched = [_RTC_REGISTERS_NUM]uint8{}

	rom_memory_path = rom_path + ".saves"
	save_needed = true
	ERAM_BANKS = make([]byte, _RAM_BANK_SIZE*headers.GetRamBankNumber())

	if headers.HasBattery() {
		loadMemory()

		savesTimer := time.Tick(time.Second)
		go func() {
			for range savesTimer {
				SaveMemory()
			}
		}()
	}
}

func loadMemory() {
	memory, err := ioutil.ReadFile(rom_memory_path)
	if err != nil || len(memory) == 0 {
		log.Printf("Error during loading previous games")
		return
	}
	ERAM_BANKS = memory
	log.Printf("!! Saves successfully loaded")
}
func SaveMemory() {
	if save_needed {
		// log.Printf("!! Saving ... \n")
		err := ioutil.WriteFile(rom_memory_path, ERAM_BANKS, 0666)
		if err != nil {
			log.Printf("Error during saving games")
			return
		}
		save_needed = false
		// log.Printf("!! Successfully saved \n")
	}

}

func WriteToRomMemory(addr uint, value uint8) {

	if addr >= 0 && addr <= _RAM_ENABLE_END {
		if value == _RAM_ENABLE_VALUE {
			ram_enabled = true
		}
		if value == _RAM_DISABLE_VALUE {
			ram_enabled = false
		}
		return
	}

	if addr >= _ROM_BANK_NUMBER_START && addr <= _ROM_BANK_NUMBER_END {
		rom_bank = uint(value) % headers.GetRomBankNumber()
		if rom_bank == 0 {
			rom_bank = 1
		}
		if headers.IsMBC1() && headers.GetRomBankBits() > 5 {
			rom_bank = ram_bank<<5 + rom_bank
		}

		return
	}

	if addr >= _RAM_BANK_NUMBER_START && addr <= _RAM_BANK_NUMBER_END {
		ram_bank = uint(value)

		if headers.IsMBC3() {
			rtc_active = value >= 0x8
			if rtc_active {
				rtc = value - 0x8
			}
		}
		return
	}

	if addr >= _BANKING_MODE_START && addr <= _BANKING_MODE_END {
		if headers.IsMBC3() {
			// check for latch : 0x0 -> 0x1
			if value == 0x1 {
				rtc_latched = true
				copy(rtc_registers_latched[:], rtc_registers[:])
			} else if value == 0x0 {
				rtc_latched = false
			}
			return
		}

		banking_mode = (value & 0b1) == 0x1
		if banking_mode == false {
			ram_bank = 0
		}
		return
	}

	log.Fatalf("Read only memory %04X\n", addr)
}

func ReadFromRomMemory(addr uint) uint8 {
	if addr <= _ROM0_END {
		return ROM[addr]
	}
	// has to be at least MBC1 to use banks
	// if !headers.IsMBC1() {
	// 	return 0xFF
	// }

	if addr >= _ROM1_START && addr <= _ROM1_END {
		return ROM[addr+(rom_bank-1)*_ROM_BANK_SIZE]
	}

	log.Fatalf("Not handled %04X (ROM)\n", addr)
	return 0
}

func WriteToRamMemory(addr uint, value uint8) {

	if addr >= _RAM_START && addr <= _RAM_END {
		// switchable bank
		if headers.IsCGB() && addr >= _RAM_CGB_START && addr <= _RAM_CGB_END {
			WRAM_CGB[ppu.GetWRAMBank()][addr] = value
			return
		}
		WRAM[addr-_RAM_START] = value
		return
	}

	// has to be at least MBC1 to use banks
	// if !headers.IsMBC1() {
	// 	return
	// }

	if addr >= _ERAM_START && addr <= _ERAM_END {
		if ram_enabled {
			if headers.IsMBC3() {
				// if rtc_active {
				if ram_bank >= 0x8 {
					rtc_registers[rtc] = value
					return
				}
			}
			off := addr - _ERAM_START + ram_bank*_RAM_BANK_SIZE
			// prev := ERAM_BANKS[off]
			ERAM_BANKS[off] = value
			save_needed = true
			// if banking_mode {
			// 	SaveMemory()
			// }
			return
		}
	}

	log.Fatalf("Not handled %04X (RAM)\n", addr)
}

func ReadFromRamMemory(addr uint) uint8 {
	if addr >= _RAM_START && addr <= _RAM_END {
		// switchable bank
		if headers.IsCGB() && addr >= _RAM_CGB_START && addr <= _RAM_CGB_END {
			return WRAM_CGB[ppu.GetWRAMBank()][addr]
		}
		return WRAM[addr-_RAM_START]
	}
	// has to be at least MBC1 to use banks
	// if !headers.IsMBC1() || !ram_enabled {
	// 	return 0xFF
	// }

	if addr >= _ERAM_START && addr <= _ERAM_END {
		if headers.IsMBC3() {
			// if rtc_active {
			if ram_bank >= 0x8 {
				if rtc_latched {
					return rtc_registers_latched[rtc]
				} else {
					return rtc_registers[rtc]
				}
			}
			return ERAM_BANKS[addr-_ERAM_START+ram_bank*_RAM_BANK_SIZE]
		}

		if ram_enabled {
			return ERAM_BANKS[addr-_ERAM_START+ram_bank*_RAM_BANK_SIZE]
		} else {
			return 0xFF
		}
	}

	log.Fatalf("Not handled %04X\n", addr)
	return 0
}

func GetRomBank() uint {
	return rom_bank
}
func GetRamBank() uint {
	return ram_bank
}
