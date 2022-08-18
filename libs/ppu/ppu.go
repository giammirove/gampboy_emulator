package ppu

import (
	"log"

	"github.com/giammirove/gbemu/libs/utility"
)

var _COLORS [4]uint32 = [4]uint32{0xFFFFFFFF, 0xFFAAAAAA, 0xFF555555, 0xFF000000}
var bg_colors [4]uint32
var obp0_colors [4]uint32
var obp1_colors [4]uint32

const _VRAM_START_ADDR = uint(0x8000)
const _VRAM_END_ADDR = uint(0x9FFF)
const _OAM_START_ADDR = uint(0xFE00)
const _OAM_END_ADDR = uint(0xFE9F)

var _VRAM [_VRAM_END_ADDR - _VRAM_START_ADDR + 1]byte
var _OAM [_OAM_END_ADDR - _OAM_START_ADDR + 1]byte

func Init() {
	InitLCD()
	InitFetcher()

	for i := 0; i < len(_COLORS); i++ {
		bg_colors[i] = _COLORS[i]
		obp0_colors[i] = _COLORS[i]
		obp1_colors[i] = _COLORS[i]
	}
}

func IsInVRAM(addr uint) bool {
	return addr >= _VRAM_START_ADDR && addr <= _VRAM_END_ADDR
}
func IsInOAM(addr uint) bool {
	return addr >= _OAM_START_ADDR && addr <= _OAM_END_ADDR
}
func IsInPPU(addr uint) bool {
	return IsInVRAM(addr) || IsInPPU(addr)
}

func ReadFromVRAMMemory(addr uint) byte {
	// mode := GetModeSTAT()
	// if mode == _MODE_PIXEL_DRAWING {
	// 	return 0xFF
	// }
	if !IsInVRAM(addr) {
		log.Fatalf("Address not in VRAM %04X\n", addr)
	}
	return _VRAM[addr-_VRAM_START_ADDR]
}
func WriteToVRAMMemory(addr uint, value byte) {
	// mode := GetModeSTAT()
	// if mode == _MODE_PIXEL_DRAWING {
	// 	return
	// }
	if !IsInVRAM(addr) {
		log.Fatalf("Address not in VRAM %04X\n", addr)
	}
	_VRAM[addr-_VRAM_START_ADDR] = value
}

func ReadFromOAMMemory(addr uint) byte {
	// mode := GetModeSTAT()
	// if mode == 2 {
	// 	return 0xFF
	// }
	if !IsInOAM(addr) {
		log.Fatalf("Address not in OAM %04X\n", addr)
	}
	return _OAM[addr-_OAM_START_ADDR]
}
func WriteToOAMMemory(addr uint, value byte) {
	// mode := GetModeSTAT()
	// if mode == 2 {
	// 	return
	// }
	if !IsInOAM(addr) {
		log.Fatalf("Address not in OAM %04X\n", addr)
	}
	_OAM[addr-_OAM_START_ADDR] = value
	// if addr == 0xFE10+3 {
	// 	log.Printf("mod to %08b\n", value)
	// }
}

func GetTileBlock(addr uint) uint {
	if addr >= _BLOCK0_START_ADDR && addr <= _BLOCK0_END_ADDR {
		return 0
	} else if addr >= _BLOCK1_START_ADDR && addr <= _BLOCK1_END_ADDR {
		return 1
	} else if addr >= _BLOCK2_START_ADDR && addr <= _BLOCK2_END_ADDR {
		return 2
	} else {
		log.Fatalf("Tile block not recognized %04X\n", addr)
		return 0
	}
}

func GetSpriteBGtoOAMPriority(addr uint) bool {
	return utility.GetBit(GetSpriteFlags(addr), _BG_OAM_PRIOR_BIT) == 0x1
}
func GetSpriteVerticalFlip(addr uint) bool {
	return utility.GetBit(GetSpriteFlags(addr), _VERTICAL_FLIP_BIT) == 0x1
}
func GetSpriteHorizontalFlip(addr uint) bool {
	return utility.GetBit(GetSpriteFlags(addr), _HORIZONTAL_FLIP_BIT) == 0x1
}
func GetSpritePaletteNumber(addr uint) bool {
	return utility.GetBit(GetSpriteFlags(addr), _PALETTE_NUMBER_BIT) == 0x1
}
func GetSpriteTileVRAMBankNumber(addr uint) bool {
	return utility.GetBit(GetSpriteFlags(addr), _TILE_VRAM_BANK_NUM_BIT) == 0x1
}
func GetSpriteBackgroundPaletteNumber(addr uint) uint {
	return GetSpriteFlags(addr) & _BACKGROUND_PALETTE_NUM_MASK & 0b111
}
func GetColor(c uint) uint32 {
	return _COLORS[c]
}
func IsTransparent(val uint) bool {
	return val == 0
}
func GetBGColor(c uint) uint32 {
	return bg_colors[c]
}
func GetOBP0Color(c uint) uint32 {
	return obp0_colors[c]
}
func GetOBP1Color(c uint) uint32 {
	return obp1_colors[c]
}

func GetSpriteYPosition(addr uint) uint {
	return uint(ReadFromOAMMemory(addr + _Y_POS_BYTE))
}
func GetSpriteXPosition(addr uint) uint {
	return uint(ReadFromOAMMemory(addr + _X_POS_BYTE))
}
func GetSpriteTileIndex(addr uint) uint {
	return uint(ReadFromOAMMemory(addr + _TILE_INDEX_BYTE))
}
func GetSpriteFlags(addr uint) uint {
	return uint(ReadFromOAMMemory(addr + _FLAGS_BYTE))
}
