package ppu

import (
	"log"

	"github.com/giammirove/gbemu/libs/interrupts"
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

var target_time uint32 = 1000 / 60
var prev_time uint32

var lcd_registers [_LCD_REGISTER_NUM]uint

func InitLCD() {
	WriteToLCDMemory(_LCDC, 0x91)
	WriteToLCDMemory(_STAT, 0x81)
	SetLY(91)
	fps = 0
}

func IsLCDAddr(addr uint) bool {
	return addr >= _REGISTER_BASE && addr <= _WX
}

func ReadFromLCDMemory(addr uint) uint {
	return lcd_registers[addr-_REGISTER_BASE]
}
func WriteToLCDMemory(addr uint, value uint) {
	if addr == _STAT {
		value = value&0b11111000 | 0x80
	}

	if addr == _DMA {
		DMAStart()
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

	prev_enable := GetLCDCEnable()
	lcd_registers[addr-_REGISTER_BASE] = value

	if addr == _LCDC {
		// resetted
		if !GetLCDCEnable() {
			lcd_registers[_STAT-_REGISTER_BASE] = getSTAT() & 0b11111100
		} else {
			// otherwise first time this check will be skipped
			updateLYFlag()
		}
		if prev_enable != GetLCDCEnable() {
			SetLY(0)
			current_dots = 0
		}
	}
}

// value is made like this : ZZWWYYXX
// where XX is 0-3 index for color 0
// etc ...
func UpdatePalette(_colors *([4]uint32), value uint) {
	for i := 0; i < len(*_colors); i++ {
		// shift value and take last two bits
		index := (value >> (i * 2)) & 0x3
		(*_colors)[i] = _COLORS[index]
	}
}

func GetCurrentFrame() int {
	return current_frame
}

var DelayGUI func(delay uint32)
var TicksGUI func() uint32
var ColorPixel func(x uint, y uint, color uint32)
var RefreshGUI func()

func LCDTick() {
	current_dots++
	if GetLCDCEnable() {
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
					fps++
					// log.Printf("update frame %d\n", current_frame)
					now := TicksGUI()
					if now-prev_time >= 1000 {
						// log.Printf("FPS %d\n", fps)
						fps = 0
						prev_time = now
						// DelayGUI(now - prev_time)
					}
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
				// if GetLY() == 3 {
				// 	log.Fatal("stop")
				// }
				// log.Println()
				FetcherClearFIFO()
				SetSTATModeHBlank()
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
}
