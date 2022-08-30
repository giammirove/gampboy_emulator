package ppu

import (
	"log"

	"github.com/giammirove/gampboy_emulator/internal/interrupts"
)

const _TOTAL_DOTS = 456
const _MODE_OAM_SCAN_DOTS = 80
const _MODE_DRAWING_DOTS = 172
const _MODE_HBLANK_DOTS = 87
const _LY_VBLANK_START = 144
const _LY_VBLANK_END = 153
const _DOTS_PER_LINE = 456
const _MODE_HBLANK = 0
const _MODE_VBLANK = 1
const _MODE_OAM = 2
const _MODE_PIXEL_DRAWING = 3

var current_dots uint
var current_frame int
var fps uint

var current_speed = 0

var target_time uint32 = 1000 / 60
var prev_time uint32

var lcd_registers [_LCD_CGB_REGISTER_NUM]uint

func InitLCD() {
	lcd_registers[_LCDC-_REGISTER_BASE] = 0x91
	lcd_registers[_STAT-_REGISTER_BASE] = 0x81
	lcd_registers[_LY-_REGISTER_BASE] = 0x91
	lcd_registers[_LYC-_REGISTER_BASE] = 0x00
	lcd_registers[_DMA-_REGISTER_BASE] = 0xFF
	lcd_registers[_BGP-_REGISTER_BASE] = 0xFC
	lcd_registers[_WY-_REGISTER_BASE] = 0x00
	lcd_registers[_WX-_REGISTER_BASE] = 0x00
	lcd_registers[_KEY1-_REGISTER_BASE] = 0xFF
	lcd_registers[_VBK-_REGISTER_BASE] = 0xFF
	lcd_registers[_RP-_REGISTER_BASE] = 0xFF
	lcd_registers[_BGPI-_REGISTER_BASE] = 0xFF
	lcd_registers[_BGPD-_REGISTER_BASE] = 0xFF
	lcd_registers[_OBPI-_REGISTER_BASE] = 0xFF
	lcd_registers[_OBPD-_REGISTER_BASE] = 0xFF
	lcd_registers[_SVBK-_REGISTER_BASE] = 0xFF
	fps = 0
	current_dots = 0
}

func IsLCDAddr(addr uint) bool {
	r := addr >= _REGISTER_BASE && addr <= _WX || (addr >= _KEY0 && addr <= _SVBK)
	return r
}

func ReadFromLCDMemory(addr uint) uint {
	if !IsLCDAddr(addr) {
		log.Fatalf("LCD address not recognized %04X\n", addr)
	}
	if addr >= _BGPI {
		addr -= _LCD_CGB_REGISTER_OFFSET
	}
	if addr == _KEY1 {
		// utility.WaitHere()
	}
	if addr == _HDMA1 || addr == _HDMA2 || addr == _HDMA3 || addr == _HDMA4 {
		// log.Printf("HDMA %04X\n", addr)
	}
	if addr == _DMA {
	}
	return lcd_registers[addr-_REGISTER_BASE]
}
func WriteToLCDMemory(addr uint, value uint) {
	if !IsLCDAddr(addr) {
		log.Fatalf("LCD address not recognized %04X\n", addr)
	}
	if addr == _STAT {
		value = value&0b11111000 | 0x80
	}

	if addr == _BGP {
		UpdatePalette(&bg_colors, value)
	}
	if addr == _OBP0 {
		UpdatePalette(&obp0_colors, value&0b11111100)
	}
	if addr == _OBP1 {
		UpdatePalette(&obp1_colors, value&0b11111100)
	}
	// if headers.IsCGB() {
	if addr == _KEY1 {
		value = value & 0b1
		// prepare speed switch
		if value == 1 {
			switchSpeed()
			value = uint(current_speed) << 7
		}
		// utility.WaitHere()
	}
	if addr == _HDMA1 || addr == _HDMA2 || addr == _HDMA3 || addr == _HDMA4 {
		// log.Printf("HDMA %04X -> %04X\n", addr, value)
	}
	if addr == _BGPD {
		UpdateCGBPalette(&cgb_bg_colors, uint32(value))
	}
	if addr == _OBPD {
		UpdateCGBPalette(&cgb_obp_colors, uint32(value))
	}
	if addr == _VBK {
		// just first bit is important, others set to 1
		value = 0xFE | (value & 0b1)
	}
	if addr == _BGPI || addr == _OBPI {
		// active the 6-th bit
		value |= 0x40
	}
	// }

	if addr >= _BGPI {
		addr -= _LCD_CGB_REGISTER_OFFSET
	}
	lcd_registers[addr-_REGISTER_BASE] = value

	if addr == _DMA {
		DMAStart()
	}
	if addr == _LCDC {
		// resetted
		if !GetLCDCEnable() {
			lcd_registers[_STAT-_REGISTER_BASE] = getSTAT() & 0b11111100
		} else {
			// otherwise first time this check will be skipped
			updateLYFlag()
		}
	}

	if addr == _HDMA5 {
		HDMAStart()
	}
}
func IsDoubleSpeed() bool {
	return current_speed == 1
}
func switchSpeed() {
	current_speed = 1 - current_speed
}

func GetCurrentFrame() int {
	return current_frame
}

var DelayGUI func(delay uint32)
var TicksGUI func() uint32
var wait uint = 0

func LCDTick() {
	if !GetLCDCEnable() {
		SetLY(0)
		current_dots = 0
		stat := getSTAT()
		stat &= 252
		setSTAT(stat)
		return
	}
	current_dots++
	mode := GetModeSTAT()
	switch mode {
	case _MODE_HBLANK:
		if current_dots >= _DOTS_PER_LINE {
			IncrementLY()
			current_dots = 0
			if GetLY() >= _LY_VBLANK_START {
				SetSTATModeVBlank()
				interrupts.RequestInterruptVBlank()
				if GetSTATINTVBlank() {
					interrupts.RequestInterruptSTAT()
				}
				current_frame++
				// ticks := TicksGUI()
				// frame_time := ticks - prev_time
				// if frame_time < target_time {
				// 	// DelayGUI(target_time - frame_time)
				// }
				// prev_time = ticks
			} else {
				SetSTATModeOAM()
				if GetSTATINTOAM() {
					interrupts.RequestInterruptSTAT()
				}
			}
		}
		break
	case _MODE_VBLANK:
		if current_dots >= _DOTS_PER_LINE {
			IncrementLY()
			current_dots = 0
			if GetLY() > _LY_VBLANK_END {
				SetLY(0)
				ResetWindowLineCounter()
				SetSTATModeOAM()
				if GetSTATINTOAM() {
					interrupts.RequestInterruptSTAT()
				}
			}
		}
		break
	case _MODE_OAM:
		if current_dots >= _MODE_OAM_SCAN_DOTS {
			FetcherStart()
			SetSTATModePixelDrawing()
		}
		if current_dots == 1 {
			FetcherOamLoad()
		}
		break
	case _MODE_PIXEL_DRAWING:
		FetcherTick()

		if FetcherGetPushedX() >= 160 {
			// DrawScanline()
			FetcherClearFIFO()
			SetSTATModeHBlank()
			HDMATransfer()
			// TODO: check if necessary
			if GetSTATINTHBlank() {
				interrupts.RequestInterruptSTAT()
			}
		}
		break
	default:
		log.Fatalf("LCD tick mode not recognized %d\n", mode)
	}
	updateLYFlag()
}
