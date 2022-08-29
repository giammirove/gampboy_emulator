package ppu

import (
	"github.com/giammirove/gampboy_emulator/internal/headers"
)

var ReadFromMemory func(addr uint, bytes ...uint) uint

var dma_old uint
var dma_delay uint
var dma_transferring bool
var dma_starting bool
var current_byte uint

var is_new_dma bool
var new_dma_delay uint
var new_current_byte uint
var new_dma_transferring bool
var new_dma_value uint
var new_dma_source uint
var new_dma_dest uint
var new_dma_len uint
var new_dma_mode uint
var new_dma_vram_bank uint

func InitDMA() {
	SetHDMA1(0xFF)
	SetHDMA2(0xFF)
	SetHDMA3(0xFF)
	SetHDMA4(0xFF)
	ResetHDM5()
}

func HDMAStart() {
	if headers.IsCGB() {
		prev_mode := new_dma_mode
		val := GetHDMA5()
		new_dma_value = val
		// TODO : check this one
		// lcd_registers[_HDMA5-_REGISTER_BASE] = 0
		new_dma_mode = (val >> 7) & 1
		// stopping hblank
		if prev_mode == 1 && new_dma_mode == 0 {
			new_dma_transferring = false
			// log.Printf("HDMA5 STOP %08b (%d)", val, new_current_byte)
			lcd_registers[_HDMA5-_REGISTER_BASE] |= 0x80
			return
		}
		// log.Printf("HDMA START %08b", val)
		new_dma_len = (val&0b1111111 + 1) * 0x10
		new_dma_vram_bank = GetVRAMBank()
		new_dma_transferring = new_dma_len > 0
		new_dma_delay = 1
		new_current_byte = 0
		is_new_dma = true
		HDMASetHighSource(uint8(GetHDMA1()))
		HDMASetLowSource(uint8(GetHDMA2()))
		HDMASetHighDestination(uint8(GetHDMA3()))
		HDMASetLowDestination(uint8(GetHDMA4()))
		new_dma_source &= 0xFFF0
		new_dma_dest &= 0x1FF0
		new_dma_dest = 0x8000 + new_dma_dest
		// log.Printf("source %04X\n", new_dma_source)
		// log.Printf("dest %04X\n", new_dma_dest)
		// log.Printf("len %d\n", new_dma_len)
		// log.Printf("general %t\n", IsGDMA())
		// log.Printf("vbank %d wbank %d\n", new_dma_vram_bank, GetWRAMBank())
		if new_dma_transferring {
			if IsGDMA() {
				GDMATransfer()
			}
			if IsHDMA() {
				HDMATransfer()
			}
		}
	}
}
func HDMASetHighSource(value uint8) {
	new_dma_source = uint(value)<<8 | new_dma_source&0xFF
}
func HDMASetLowSource(value uint8) {
	new_dma_source = uint(value) | new_dma_source&0xFF00
}
func HDMASetHighDestination(value uint8) {
	new_dma_dest = uint(value)<<8 | new_dma_dest&0xFF
}
func HDMASetLowDestination(value uint8) {
	new_dma_dest = uint(value) | new_dma_dest&0xFF00
}
func IsGDMA() bool {
	return new_dma_mode&1 == 0x0
}
func IsHDMA() bool {
	return new_dma_mode&1 == 0x1
}
func IsGDMATransferring() bool {
	return new_dma_transferring && IsGDMA() && headers.IsCGB()
}
func IsHDMATransferring() bool {
	return new_dma_transferring && IsHDMA() && headers.IsCGB()
}

func HDMATransfer() {
	if IsHDMATransferring() {
		if new_dma_delay > 0 {
			new_dma_delay--
			return
		}
		if GetVRAMBank() != new_dma_vram_bank {
			return
		}
		for i := 0; i < 0x10; i++ {
			b := byte(ReadFromMemory(new_dma_source + new_current_byte))

			WriteToVRAMMemory(new_dma_dest+new_current_byte, b, new_dma_vram_bank)

			new_current_byte++
			updateHDMARegisters()

			new_dma_transferring = new_current_byte < new_dma_len
			if !new_dma_transferring {
				ResetHDM5()
				return
			}
		}

	}
}
func GDMATransfer() {
	if IsGDMATransferring() {
		if new_dma_delay > 0 {
			new_dma_delay--
			return
		}
		if GetVRAMBank() != new_dma_vram_bank {
			return
		}
		b := byte(ReadFromMemory(new_dma_source + new_current_byte))

		WriteToVRAMMemory(new_dma_dest+new_current_byte, b, new_dma_vram_bank)

		new_current_byte++
		updateHDMARegisters()

		new_dma_transferring = new_current_byte < new_dma_len
		if !new_dma_transferring {
			ResetHDM5()
		}
	}
}
func updateHDMARegisters() {
	src := new_dma_source + new_current_byte
	dst := new_dma_dest + new_current_byte
	leng := int((new_dma_len-new_current_byte)/0x10 - 1)
	if leng < 0 {
		leng = 0
	}
	SetHDMA1(src & 0xFF00 >> 8)
	SetHDMA2(src & 0xFF)
	SetHDMA3(dst & 0xFF00 >> 8)
	SetHDMA4(dst & 0xFF)
	UpdateDMALength(uint(leng))
}

func DMAStart() {
	dma_old = GetDMA() << 8
	// max is 0xDF00 so in case just decrease
	if dma_old > 0xDF00 {
		dma_old -= 0x2000
	}
	dma_starting = false
	dma_delay = 2
	// log.Printf("DMA START %04X", dma_old)
}
func DMATransfer() {
	if !dma_transferring {
		return
	}
	b := byte(ReadFromMemory(dma_old + current_byte))
	WriteToOAMMemory(0xFE00+current_byte, b)
	// log.Printf("[%d] dma write to %04X (%04X)", current_byte, 0xFE00+current_byte, b)

	lcd_registers[_DMA-_REGISTER_BASE] = (dma_old + current_byte) >> 8

	current_byte++

	// since dma transfers to 0xFE00 - 0xFE9F ( so 0x9F bytes )
	dma_transferring = current_byte <= 0x9F
	if !dma_transferring {
		// log.Printf("DMA END")
	}
}

func DMATick() {
	if dma_delay > 0 {
		dma_delay--
		if dma_delay == 0 {
			dma_starting = true
		}
	}
	if dma_starting {
		dma_starting = false
		current_byte = 0
		dma_transferring = true
		// log.Printf("DMA TRANSFERRING")
	}

	DMATransfer()
}
func GetDMATransferring() bool {
	return dma_transferring
}
