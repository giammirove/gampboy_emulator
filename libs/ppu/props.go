package ppu

import (
	"github.com/giammirove/gbemu/libs/interrupts"
	"github.com/giammirove/gbemu/libs/utility"
)

/**
LCD registers are
 0 -  FF40 - LCDC
 1 -  FF41 - STAT
 2 -  FF42 - SCY
 3 -  FF43 - SCX
 4 -  FF44 - LY
 5 -  FF45 - LYC
 6 -  FF46 - DMA
 7 -  FF47 - BGP
 8 -  FF48 - OBP0
 9 -  FF49 - OBP1
 10 -  FF4A - WY
 11 - FF4B - WX
*/

const _LCDC = 0xFF40
const _STAT = 0xFF41
const _SCY = 0xFF42
const _SCX = 0xFF43
const _LY = 0xFF44
const _LYC = 0xFF45
const _DMA = 0xFF46
const _BGP = 0xFF47
const _OBP0 = 0xFF48
const _OBP1 = 0xFF49
const _WY = 0xFF4A
const _WX = 0xFF4B

const _REGISTER_BASE = _LCDC
const _END_ADDR = _WX
const _LCD_REGISTER_NUM = _END_ADDR - _REGISTER_BASE + 1

const _STAT_MODE_MASK = uint(0x3)
const _STAT_INT_LY_BIT = 6
const _STAT_INT_OAM_BIT = 5
const _STAT_INT_VBLANK_BIT = 4
const _STAT_INT_HBLANK_BIT = 3
const _STAT_LY_FLAG_BIT = 2

const _LCDC_BG_DISPLAY_BIT = 0
const _LCDC_OBJ_DISPLAY_BIT = 1
const _LCDC_OBJ_SIZE_BIT = 2
const _LCDC_BG_TILE_MAP_DISPLAY_BIT = 3
const _LCDC_BG_TILE_DATA_BIT = 4
const _LCDC_WIN_DISPLAY_BIT = 5
const _LCDC_WIN_TILE_MAP_DISPLAY_BIT = 6
const _LCDC_LCD_ENABLE_BIT = 7

const _Y_POS_BYTE = 0
const _X_POS_BYTE = 1
const _TILE_INDEX_BYTE = 2
const _FLAGS_BYTE = 3

const _BG_OAM_PRIOR_BIT = 7
const _VERTICAL_FLIP_BIT = 6
const _HORIZONTAL_FLIP_BIT = 5
const _PALETTE_NUMBER_BIT = 4
const _TILE_VRAM_BANK_NUM_BIT = 3
const _BACKGROUND_PALETTE_NUM_MASK = 0x7

const _BLOCK0_START_ADDR = 0x8000
const _BLOCK0_END_ADDR = 0x87FF
const _BLOCK1_START_ADDR = 0x8800
const _BLOCK1_END_ADDR = 0x8FFF
const _BLOCK2_START_ADDR = 0x9000
const _BLOCK2_END_ADDR = 0x97FF

const _TILE_NUM = 384
const _TILE_BYTES = 16
const _TILE_W = 8
const _TILE_H = 8

func getLCDC() uint {
	return ReadFromLCDMemory(_LCDC)
}

func GetLCDCEnable() bool {
	return utility.GetBit(getLCDC(), _LCDC_LCD_ENABLE_BIT) == 0x1
}
func GetLCDCWinTileMapDisplay() bool {
	return utility.GetBit(getLCDC(), _LCDC_WIN_TILE_MAP_DISPLAY_BIT) == 0x1
}
func GetLCDCWinDisplay() bool {
	return utility.GetBit(getLCDC(), _LCDC_WIN_DISPLAY_BIT) == 0x1
}
func GetLCDCBGWinTileDataArea() bool {
	return utility.GetBit(getLCDC(), _LCDC_BG_TILE_DATA_BIT) == 0x1
}
func GetLCDCBGTileMapDisplayArea() bool {
	return utility.GetBit(getLCDC(), _LCDC_BG_TILE_MAP_DISPLAY_BIT) == 0x1
}
func GetLCDCOBJSize() bool {
	return utility.GetBit(getLCDC(), _LCDC_OBJ_SIZE_BIT) == 0x1
}
func GetLCDCOBJDisplay() bool {
	return utility.GetBit(getLCDC(), _LCDC_OBJ_DISPLAY_BIT) == 0x1
}
func GetLCDCBGWinDisplay() bool {
	return utility.GetBit(getLCDC(), _LCDC_BG_DISPLAY_BIT) == 0x1
}

func getSTAT() uint {
	return ReadFromLCDMemory(_STAT)
}
func setSTAT(val uint) {
	lcd_registers[_STAT-_REGISTER_BASE] = val
}
func GetModeSTAT() uint {
	return getSTAT() & _STAT_MODE_MASK
}
func setSTATMode(mode uint) {
	// clear the previous one
	_stat := getSTAT()
	_stat &= ^_STAT_MODE_MASK
	_stat |= mode
	setSTAT(_stat)
}
func resetSTATMode() {
	// the first one to be executed
	SetSTATModeOAM()
}
func SetSTATModeVBlank() {
	setSTATMode(_MODE_VBLANK)
}
func SetSTATModeHBlank() {
	setSTATMode(_MODE_HBLANK)
}
func SetSTATModeOAM() {
	setSTATMode(_MODE_OAM)
}
func SetSTATModePixelDrawing() {
	setSTATMode(_MODE_PIXEL_DRAWING)
}
func GetSTATINTLY() bool {
	return utility.GetBit(getSTAT(), _STAT_INT_LY_BIT) == 0x1
}
func SetSTATINTLY(val uint) {
	setSTAT(utility.WriteBit(getSTAT(), _STAT_INT_LY_BIT, val))
}
func GetSTATINTOAM() bool {
	return utility.GetBit(getSTAT(), _STAT_INT_OAM_BIT) == 0x1
}
func SetSTATINTOAM(val uint) {
	setSTAT(utility.WriteBit(getSTAT(), _STAT_INT_OAM_BIT, val))
}
func GetSTATINTVBlank() bool {
	return utility.GetBit(getSTAT(), _STAT_INT_VBLANK_BIT) == 0x1
}
func SetSTATINTVBlank(val uint) {
	setSTAT(utility.WriteBit(getSTAT(), _STAT_INT_VBLANK_BIT, val))
}
func GetSTATINTHBlank() bool {
	return utility.GetBit(getSTAT(), _STAT_INT_HBLANK_BIT) == 0x1
}
func SetSTATINTHBlank(val uint) {
	setSTAT(utility.WriteBit(getSTAT(), _STAT_INT_HBLANK_BIT, val))
}
func GetSTATLYFlag() bool {
	return utility.GetBit(getSTAT(), _STAT_LY_FLAG_BIT) == 0x1
}
func SetSTATLYFlag(val uint) {
	setSTAT(utility.WriteBit(getSTAT(), _STAT_LY_FLAG_BIT, val))
}

func GetSCY() uint {
	return ReadFromLCDMemory(_SCY)
}
func SetSCY(val uint) {
	WriteToLCDMemory(_SCY, val)
}
func GetSCX() uint {
	return ReadFromLCDMemory(_SCX)
}
func SetSCX(val uint) {
	WriteToLCDMemory(_SCX, val)
}
func GetLY() uint {
	return ReadFromLCDMemory(_LY)
}
func SetLY(val uint) {
	WriteToLCDMemory(_LY, val)
}
func IncrementLY() {
	SetLY(GetLY() + 1)
	IncrementWindowLineCounter()
}
func updateLYFlag() {
	prev := GetSTATLYFlag()
	if GetLY() == GetLYC() {
		// interrupt
		SetSTATLYFlag(1)
		// if interrupt is enabled
		if GetSTATINTLY() && !prev {
			interrupts.RequestInterruptSTAT()
			prev = GetLY() == GetLYC()
		}
	} else {
		SetSTATLYFlag(0)
	}
}
func GetLYC() uint {
	return ReadFromLCDMemory(_LYC)
}
func SetLYC(val uint) {
	WriteToLCDMemory(_LYC, val)
}
func GetDMA() uint {
	return ReadFromLCDMemory(_DMA)
}
func SetDMA(val uint) {
	WriteToLCDMemory(_DMA, val)
}
func GetBGP() uint {
	return ReadFromLCDMemory(_BGP)
}
func SetBGP(val uint) {
	WriteToLCDMemory(_BGP, val)
}
func GetOBP0() uint {
	return ReadFromLCDMemory(_OBP0)
}
func SetOBP0(val uint) {
	WriteToLCDMemory(_OBP0, val)
}
func GetOBP1() uint {
	return ReadFromLCDMemory(_OBP1)
}
func SetOBP1(val uint) {
	WriteToLCDMemory(_OBP1, val)
}
func GetWY() uint {
	return ReadFromLCDMemory(_WY)
}
func SetWY(val uint) {
	WriteToLCDMemory(_WY, val)
}
func GetWX() uint {
	return ReadFromLCDMemory(_WX)
}
func SetWX(val uint) {
	WriteToLCDMemory(_WX, val)
}
