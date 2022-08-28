package interrupts

import (
	"log"

	"github.com/giammirove/gbemu/internal/registers"
	"github.com/giammirove/gbemu/internal/utility"
)

/**
Bit 0: VBlank   Interrupt Enable  (INT $40)  (1=Enable)
Bit 1: LCD STAT Interrupt Enable  (INT $48)  (1=Enable)
Bit 2: Timer    Interrupt Enable  (INT $50)  (1=Enable)
Bit 3: Serial   Interrupt Enable  (INT $58)  (1=Enable)
Bit 4: Joypad   Interrupt Enable  (INT $60)  (1=Enable)
*/

var _IME_enable uint
var _IME_pending uint
var _IE_REG uint
var _IF_REG uint

const _IE_ADDR = 0xFFFF
const _IF_ADDR = 0xFF0F

const _INT_NUM = 5

const _VBLANK = 0
const _LCD_STAT = 1
const _TIMER = 2
const _SERIAL = 3
const _JOYPAD = 4

const _VBLANK_ADDR = 0x40
const _LCD_STAT_ADDR = 0x48
const _TIMER_ADDR = 0x50
const _SERIAL_ADDR = 0x58
const _JOYPAD_ADDR = 0x60

var _INT [_INT_NUM][2]uint

var StackPush func(value uint, bytes ...uint)

func Init() {
	_IME_enable = 0
	_IE_REG = 0
	_IF_REG = 0xE1
	_INT = [_INT_NUM][2]uint{
		{_VBLANK, _VBLANK_ADDR},
		{_LCD_STAT, _LCD_STAT_ADDR},
		{_TIMER, _TIMER_ADDR},
		{_SERIAL, _SERIAL_ADDR},
		{_JOYPAD, _JOYPAD_ADDR},
	}
}

var Cycle func(val uint)

func HandleInterrupts() bool {
	// order matter
	for i := 0; i < _INT_NUM; i++ {
		if checkInterrupt(_INT[i][1], _INT[i][0]) {
			return true
		}
	}
	return false
}

func checkInterrupt(addr uint, bit uint) bool {
	if GetBitIE(bit) && GetBitIF(bit) {
		// if cpu is halted it takes 4 cycle more
		if GetHalted() {
			return true
		}
		return handleInterrupt(addr, bit)
	}
	return false
}

var GetHalted func() bool
var MMUWriteToMemory func(addr uint, value uint, bytes ...uint)

func handleInterrupt(addr uint, bit uint) bool {
	// prev := registers.SP()
	hi, low := utility.GetHiLow(uint16(registers.PC()))
	Cycle(4)
	registers.DecrementSP()
	Cycle(4)
	MMUWriteToMemory(registers.SP(), uint(hi))
	// check if IE of IF have been overwritten
	if !(GetBitIE(bit) && GetBitIF(bit)) {
		// TOOD: check this one
		registers.SetPC(0x0000)
		return false
	}
	Cycle(4)
	registers.DecrementSP()
	Cycle(4)
	MMUWriteToMemory(registers.SP(), uint(low))
	SetIME(0)
	DisableBitIF(bit)
	Cycle(4)
	registers.SetPC(addr)
	return true
}

func RequestInterruptTimer() {
	SetBitIF(_TIMER)
}
func RequestInterruptSTAT() {
	SetBitIF(_LCD_STAT)
}
func RequestInterruptVBlank() {
	SetBitIF(_VBLANK)
}
func RequestInterruptSerial() {
	SetBitIF(_SERIAL)
}
func RequestInterruptJoypad() {
	SetBitIF(_JOYPAD)
}

// interrupts to be enable need one cpu cycle
func EnableInterrupts() {
	_IME_pending = 1
}
func EnableInterruptsImmediately() {
	_IME_pending = 0
	_IME_enable = 1
}

func DisableInterrupts() {
	_IME_enable = 0
	_IME_pending = 0
}

func IsInterruptAddr(addr uint) bool {
	return addr == _IE_ADDR || addr == _IF_ADDR
}

func ReadFromMemory(addr uint) uint {
	if addr == _IE_ADDR {
		return GetIE()
	} else if addr == _IF_ADDR {
		return GetIF()
	}

	log.Fatalf("Invalid interrupt address (Read) (0x%08X)\n", addr)
	return 0
}
func WriteToMemory(addr uint, value uint) {
	if addr == _IE_ADDR {
		SetIE(value)
		// log.Printf("IE -> %04X\n", value)
		// utility.WaitHere()
		return
	} else if addr == _IF_ADDR {
		SetIF(value)
		// log.Printf("IF -> %04X\n", value)
		// utility.WaitHere()
		return
	}

	log.Fatalf("Invalid interrupt address (Write) (0x%08X)\n", addr)
}

func GetIME() bool {
	return _IME_enable == 1
}
func SetIME(val uint) {
	_IME_enable = val
}
func GetPendingIME() bool {
	return _IME_pending == 1
}
func SetPendingIME(val uint) {
	_IME_pending = val
}

func GetIE() uint {
	return _IE_REG
}
func GetIF() uint {
	return _IF_REG
}
func SetIE(value uint) {
	_IE_REG = value | 0b11100000
}
func SetIF(value uint) {
	_IF_REG = value | 0b11100000
}

func GetBitIE(bit uint) bool {
	return utility.GetBit(_IE_REG, bit) == 1
}
func GetBitIF(bit uint) bool {
	return utility.GetBit(_IF_REG, bit) == 1
}
func SetBitIE(bit uint) {
	SetIE(utility.SetBit(_IE_REG, bit))
}
func SetBitIF(bit uint) {
	SetIF(utility.SetBit(_IF_REG, bit))
}
func DisableBitIE(bit uint) {
	SetIE(utility.ClearBit(_IE_REG, bit))
}
func DisableBitIF(bit uint) {
	SetIF(utility.ClearBit(_IF_REG, bit))
}

func GetVBlankIE() bool {
	return utility.GetBit(_IE_REG, _VBLANK) == 1
}
func EnableVBlankIE() {
	SetBitIE(_VBLANK)
}
func DisableVBlankIE() {
	DisableBitIE(_VBLANK)
}
func GetVBlankIF() bool {
	return utility.GetBit(_IF_REG, _VBLANK) == 1
}
func EnableVBlankIF() {
	SetBitIF(_VBLANK)
}
func DisableVBlankIF() {
	DisableBitIF(_VBLANK)
}

func GetLCDSTATIE() bool {
	return utility.GetBit(_IE_REG, _LCD_STAT) == 1
}
func EnableLCDSTATIE() {
	SetBitIE(_LCD_STAT)
}
func DisableLCDSTATIE() {
	DisableBitIE(_LCD_STAT)
}
func GetLCDSTATIF() bool {
	return utility.GetBit(_IF_REG, _LCD_STAT) == 1
}
func EnableLCDSTATIF() {
	SetBitIF(_LCD_STAT)
}
func DisableLCDSTATIF() {
	DisableBitIF(_LCD_STAT)
}

func GetTimerIE() bool {
	return utility.GetBit(_IE_REG, _TIMER) == 1
}
func EnableTimerIE() {
	SetBitIE(_TIMER)
}
func DisableTimerIE() {
	DisableBitIE(_TIMER)
}
func GetTimerIF() bool {
	return utility.GetBit(_IF_REG, _TIMER) == 1
}
func EnableTimerIF() {
	SetBitIF(_TIMER)
}
func DisableTimerIF() {
	DisableBitIF(_TIMER)
}

func GetSerialIE() bool {
	return utility.GetBit(_IE_REG, _SERIAL) == 1
}
func EnableSerialIE() {
	SetBitIE(_SERIAL)
}
func DisableSerialIE() {
	DisableBitIE(_SERIAL)
}
func GetSerialIF() bool {
	return utility.GetBit(_IF_REG, _SERIAL) == 1
}
func EnableSerialIF() {
	SetBitIF(_SERIAL)
}
func DisableSerialIF() {
	DisableBitIF(_SERIAL)
}

func GetJoypadIE() bool {
	return utility.GetBit(_IE_REG, _JOYPAD) == 1
}
func EnableJoypadIE() {
	SetBitIE(_JOYPAD)
}
func DisableJoypadIE() {
	DisableBitIE(_JOYPAD)
}
func GetJoypadIF() bool {
	return utility.GetBit(_IF_REG, _JOYPAD) == 1
}
func EnableJoypadIF() {
	SetBitIF(_JOYPAD)
}
func DisableJoypadIF() {
	DisableBitIF(_JOYPAD)
}
