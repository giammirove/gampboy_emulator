package utility

import (
	"bufio"
	"log"
	"os"
)

const _MASK_LOW = 0xFF
const _MASK_HI = 0xFF00
const _HALF_WORD = 8

func WaitHere(msg ...string) {
	if len(msg) > 0 {
		log.Printf("[...] %s\n", msg[0])
	}
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func SetHiLow(hi uint8, low uint8) uint16 {
	return uint16(hi)<<_HALF_WORD | uint16(low)
}

func GetHiLow(val uint16) (uint8, uint8) {
	low := uint8(val & _MASK_LOW)
	hi := uint8(val & _MASK_HI >> _HALF_WORD)
	return hi, low
}

func GetBit(addr uint, bit uint) uint {
	return (addr & (1 << bit)) >> bit & 1
}
func TestBit(addr uint, bit uint) bool {
	return ((addr & (1 << bit)) >> bit & 1) == 0x1
}

func SetBit(addr uint, bit uint) uint {
	return (1 << bit) | addr
}
func ClearBit(addr uint, bit uint) uint {
	return addr & ^(1 << bit)
}
func WriteBit(addr uint, bit uint, value uint) uint {
	value = value & 0x1
	if value == 0x0 {
		return ClearBit(addr, bit)
	}
	return SetBit(addr, bit)
}

func CheckCarry(n1 uint, n2 uint) bool {
	return ((n1&0xff)+(n2&0xff) > 0xFF)
}
func CheckTriCarry(n1 uint, n2 uint, n3 uint) bool {
	return ((n1&0xff)+(n2&0xff)+(n3&0xff) > 0xFF)
}
func CheckCarry16bit(n1 uint, n2 uint) bool {
	return (uint32(n1)+uint32(n2) > 0xFFFF)
}
func CheckHalfCarry(n1 uint, n2 uint) bool {
	return ((n1 & 0xf) + (n2 & 0xf)) > 0xF
}

// registers.A()&0xF+uint(value)&0xF+c > 0xF
func CheckTriHalfCarry(n1 uint, n2 uint, n3 uint) bool {
	return ((n1 & 0xf) + (n2 & 0xf) + (n3 & 0xf)) > 0xF
}
func CheckHalfCarry16bit(n1 uint, n2 uint) bool {
	// return (((n1 & 0xFFF) + (n2 & 0xFFF)) & 0x1000) > 0xFFF
	return ((n1 & 0xFFF) + (n2 & 0xFFF)) > 0xFFF
}

func CheckCarrySub(n1 uint, n2 uint) bool {
	return (int(n1&0xFF) - int(n2&0x0FF)) < 0
}
func CheckTriCarrySub(n1 uint, n2 uint, n3 uint) bool {
	return (int(n1&0xFF) - int(n2&0x0FF) - int(n3&0x0FF)) < 0
}
func CheckHalfCarrySub(n1 uint, n2 uint) bool {
	return (int(n1&0x0F) - int(n2&0x0F)) < 0
}
func CheckTriHalfCarrySub(n1 uint, n2 uint, n3 uint) bool {
	return (int(n1&0x0F) - int(n2&0x0F) - int(n3&0x0F)) < 0
}

func Swap8bit(val uint8) uint8 {
	return ((val&0x0F)<<4 | (val&0xF0)>>4)
}

func Contains(arr []uint, val uint) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i] == val {
			return true
		}
	}
	return false
}
