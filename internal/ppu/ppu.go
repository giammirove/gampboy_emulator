package ppu

import (
	"log"

	"github.com/giammirove/gampboy_emulator/internal/utility"
)

var _COLORS [4]uint32 = [4]uint32{0xFFFFFFFF, 0xFFAAAAAA, 0xFF555555, 0xFF000000}
var bg_colors [4]uint32
var obp0_colors [4]uint32
var obp1_colors [4]uint32
var cgb_bg_colors [32]uint32
var cgb_obp_colors [32]uint32

const _VRAM_START_ADDR = uint(0x8000)
const _VRAM_END_ADDR = uint(0x9FFF)
const _OAM_START_ADDR = uint(0xFE00)
const _OAM_END_ADDR = uint(0xFE9F)

var _VRAM [2][_VRAM_END_ADDR - _VRAM_START_ADDR + 1]byte
var _OAM [_OAM_END_ADDR - _OAM_START_ADDR + 1]byte

func Init() {
	InitLCD()
	InitDMA()
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

func CanAccessVRAM() bool {
	return GetModeSTAT() != _MODE_PIXEL_DRAWING
}
func ReadFromVRAMMemory(addr uint, bank ...uint) byte {
	// mode := GetModeSTAT()
	// if mode == _MODE_PIXEL_DRAWING {
	// 	return 0xFF
	// }
	if !IsInVRAM(addr) {
		log.Fatalf("Address not in VRAM %04X\n", addr)
	}
	b := GetVRAMBank()
	if len(bank) == 1 {
		b = bank[0]
	}
	return _VRAM[b][addr-_VRAM_START_ADDR]
}
func WriteToVRAMMemory(addr uint, value byte, bank ...uint) {
	mode := GetModeSTAT()
	if mode == _MODE_PIXEL_DRAWING {
		// return
	}
	if !IsInVRAM(addr) {
		log.Fatalf("Address not in VRAM %04X\n", addr)
	}
	b := GetVRAMBank() & 1
	if len(bank) == 1 {
		b = bank[0]
	}
	_VRAM[b][addr-_VRAM_START_ADDR] = value
}

func CanAccessOAM() bool {
	return GetModeSTAT() == _MODE_HBLANK || GetModeSTAT() == _MODE_VBLANK
}
func ReadFromOAMMemory(addr uint) byte {
	if !IsInOAM(addr) {
		log.Fatalf("Address not in OAM %04X (Read)\n", addr)
	}
	if addr == 0xFE00 {
		// utility.WaitHere("writing to FE00")
	}
	return _OAM[addr-_OAM_START_ADDR]
}
func WriteToOAMMemory(addr uint, value byte) {
	// mode := GetModeSTAT()
	// if mode == 2 {
	// 	return
	// }
	if !IsInOAM(addr) {
		log.Fatalf("Address not in OAM %04X (Write)\n", addr)
	}
	_OAM[addr-_OAM_START_ADDR] = value
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

// Non CGB mode only
func GetSpritePaletteNumber(addr uint) bool {
	return utility.GetBit(GetSpriteFlags(addr), _PALETTE_NUMBER_BIT) == 0x1
}

// CGB mode only
func GetSpriteTileVRAMBankNumber(addr uint) bool {
	return utility.GetBit(GetSpriteFlags(addr), _TILE_VRAM_BANK_NUM_BIT) == 0x1
}

// CGB mode only
func GetSpriteCGBPaletteNumber(addr uint) uint {
	return GetSpriteFlags(addr) & _CGB_PALETTE_NUM_MASK
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

func GetCGBBGColor(tile_addr uint, index uint) uint32 {
	// if !CanAccessVRAM() {
	// 	return 0xFFFFFFFF
	// }
	return adjustColor(cgb_bg_colors[GetCGBBGPaletteNumber(tile_addr)*4+index])
}
func GetCGBBGPaletteNumber(tile_addr uint) uint {
	return uint(ReadFromVRAMMemory(tile_addr, 1)) & _CGB_PALETTE_NUM_MASK
}
func GetCGBOBPColor(obp_addr uint, index uint) uint32 {
	// if !CanAccessVRAM() {
	// 	return 0xFFFFFFFF
	// }
	return adjustColor(cgb_obp_colors[GetSpriteCGBPaletteNumber(obp_addr)*4+index])
}
func adjustColor(color uint32) uint32 {

	blue := color & 0x1F
	green := (color >> 8) & 0x1F
	red := (color >> 16) & 0x1F

	new_red := (red*32 + green*0 + blue*0) >> 2
	new_green := (green*32 + blue*0) >> 2
	new_blue := (red*0 + green*0 + blue*32) >> 2

	return new_red<<16 | new_green<<8 | new_blue
}

func min(a uint32, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

// value is made like this : ZZWWYYXX
// where XX is 0-3 index for color 0
// etc ...
func UpdatePalette(_colors *([4]uint32), value uint) {
	if !CanAccessVRAM() {
		return
	}
	for i := 0; i < len(*_colors); i++ {
		// shift value and take last two bits
		index := (value >> (i * 2)) & 0x3
		(*_colors)[i] = _COLORS[index]
	}
}

func UpdateCGBPalette(_colors *([32]uint32), value uint32) {
	if !CanAccessVRAM() {
		return
	}
	addr := GetBGPIAddress()
	if _colors == &cgb_obp_colors {
		addr = GetOBPIAddress()
	}
	// coloring lower byte
	if addr&1 == 0 {
		r := value & 0b11111
		g := (value & 0b11100000) >> 5
		(*_colors)[addr/2] = uint32(r<<16 | g<<8)
	} else {
		g := (value&0b11)<<3 | ((*_colors)[addr/2] & 0xFF00 >> 8)
		b := (value&0b01111100)>>2 | (*_colors)[addr/2]&0xFF
		(*_colors)[addr/2] |= g<<8 | b
	}
	if _colors == &cgb_bg_colors && GetBGPIAutoIncrement() {
		IncBGPI()
	}
	if _colors == &cgb_obp_colors && GetOBPIAutoIncrement() {
		IncOBPI()
	}
}
func GetCGBBGPriority(tile_addr uint) bool {
	return utility.GetBit(uint(ReadFromVRAMMemory(tile_addr, 1)), _BG_OAM_PRIOR_BIT) == 0x1
}
func GetCGBBGVRAMBank(tile_addr uint) bool {
	return utility.GetBit(uint(ReadFromVRAMMemory(tile_addr, 1)), _TILE_VRAM_BANK_NUM_BIT) == 0x1
}
func GetCGBBGVerticalFlip(tile_addr uint) bool {
	return utility.GetBit(uint(ReadFromVRAMMemory(tile_addr, 1)), _VERTICAL_FLIP_BIT) == 0x1
}
func GetCGBBGHorizontalFlip(tile_addr uint) bool {
	return utility.GetBit(uint(ReadFromVRAMMemory(tile_addr, 1)), _HORIZONTAL_FLIP_BIT) == 0x1
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
