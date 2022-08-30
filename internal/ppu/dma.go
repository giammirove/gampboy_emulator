package ppu

import (
	"github.com/giammirove/gampboy_emulator/internal/headers"
)

var ReadFromMemory func(addr uint, bytes ...uint) uint

var dma_old uint
var dma_delay uint
var dma_transferring bool
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
var new_dma_wram_bank uint

func InitDMA() {
	SetHDMA1(0xFF)
	SetHDMA2(0xFF)
	SetHDMA3(0xFF)
	SetHDMA4(0xFF)
	ResetHDM5()
}

func HDMAStart() {
	if headers.IsCGB() {
		val := GetHDMA5()
		// TODO : check this one
		// lcd_registers[_HDMA5-_REGISTER_BASE] = 0
		tmp_dma_mode := (val >> 7) & 1
		if new_dma_mode == 1 && !new_dma_transferring {
			lcd_registers[_HDMA5-_REGISTER_BASE] &= 0x7F
		} else if new_dma_mode == 1 && tmp_dma_mode == 0 && new_dma_transferring {
			new_dma_transferring = false
			// fmt.Printf("HDMA STOP %08b\n", val)
			lcd_registers[_HDMA5-_REGISTER_BASE] |= 0x80
			return
		}
		new_dma_value = val
		new_dma_mode = tmp_dma_mode
		new_dma_len = (val&0x7F + 1) * 0x10
		new_dma_vram_bank = GetVRAMBank()
		new_dma_wram_bank = GetWRAMBank()
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
		// if new_dma_mode == 1 {
		// 	fmt.Printf("HDMA START %08b\n", val)
		// } else {
		// 	fmt.Printf("GDMA START %08b\n", val)
		// }
		// fmt.Printf("%04X -> %04X (%d)\n", new_dma_source, new_dma_dest, new_dma_len)
		// if new_dma_len == 1024 {
		// 	utility.WaitHere()
		// }
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
		if GetVRAMBank() != new_dma_vram_bank {
			return
		}
		for i := 0; i < 0x10; i++ {
			newDMATransfer()
		}

	}
}
func GDMATransfer() {
	if IsGDMATransferring() {
		newDMATransfer()
	}
}
func newDMATransfer() {
	b := byte(ReadFromMemory(new_dma_source + new_current_byte))

	WriteToVRAMMemory(new_dma_dest+new_current_byte, b)

	new_current_byte++
	updateHDMARegisters()

	new_dma_transferring = new_current_byte < new_dma_len
	if !new_dma_transferring {
		ResetHDM5()
	}
}
func updateHDMARegisters() {
	src := new_dma_source + new_current_byte
	dst := new_dma_dest + new_current_byte
	leng := int((new_dma_len-new_current_byte)/0x10 - 1)
	if leng < 0 {
		leng = 0
	}
	SetHDMA1(src >> 8)
	SetHDMA2(src & 0xFF)
	SetHDMA3(dst >> 8)
	SetHDMA4(dst & 0xFF)
	UpdateDMALength(uint(leng))
}

func DMAStart() {
	dma_old = GetDMA() << 8
	// max is 0xDF00 so in case just decrease
	if dma_old > 0xDF00 {
		dma_old -= 0x2000
	}
	dma_delay = 2
	// log.Printf("DMA START %04X", dma_old)
}
func DMATransfer() {
	if !dma_transferring {
		return
	}
	b := byte(ReadFromMemory(dma_old + current_byte))
	WriteToOAMMemory(0xFE00+current_byte, b)

	lcd_registers[_DMA-_REGISTER_BASE] = (dma_old + current_byte) >> 8

	current_byte++

	// since dma transfers to 0xFE00 - 0xFE9F ( so 0x9F bytes )
	dma_transferring = current_byte <= 0x9F
}

func DMATick() {
	if dma_delay > 0 {
		dma_delay--
		if dma_delay == 0 {
			current_byte = 0
			dma_transferring = true
		}
	}

	DMATransfer()
}
func GetDMATransferring() bool {
	return dma_transferring
}
