package joypad

import (
	"log"

	"github.com/giammirove/gampboy_emulator/internal/interrupts"
	"github.com/giammirove/gampboy_emulator/internal/utility"
)

const _JOYPAD_ADDR = 0xFF00
const _JOYPAD_ACTIONS_BIT = 5
const _JOYPAD_DIRECTIONS_BIT = 4
const _JOYPAD_DOWN_START_BIT = 3
const _JOYPAD_UP_SELECT_BIT = 2
const _JOYPAD_LEFT_B_BIT = 1
const _JOYPAD_RIGHT_A_BIT = 0

var actions uint
var directions uint
var btn_start uint
var btn_select uint
var btn_b uint
var btn_a uint
var btn_down uint
var btn_up uint
var btn_left uint
var btn_right uint

func Init() {
	// default is all 1 == all not pressed
	Reset()
}

func Reset() {
	actions = 1
	directions = 1
	btn_start = 1
	btn_select = 1
	btn_a = 1
	btn_b = 1
	btn_down = 1
	btn_up = 1
	btn_left = 1
	btn_right = 1
}

func IsJoypadAddr(addr uint) bool {
	return addr == _JOYPAD_ADDR
}

func ReadFromMemory(addr uint) uint {
	if addr != _JOYPAD_ADDR {
		log.Fatalf("Invalid sound address (0x%8X)", addr)
	}
	r := uint(0xCF)
	if actions == 0x0 {
		r = utility.WriteBit(r, _JOYPAD_DOWN_START_BIT, btn_start)
		r = utility.WriteBit(r, _JOYPAD_UP_SELECT_BIT, btn_select)
		r = utility.WriteBit(r, _JOYPAD_LEFT_B_BIT, btn_b)
		r = utility.WriteBit(r, _JOYPAD_RIGHT_A_BIT, btn_a)
	}
	if directions == 0x0 {
		r = utility.WriteBit(r, _JOYPAD_DOWN_START_BIT, btn_down)
		r = utility.WriteBit(r, _JOYPAD_UP_SELECT_BIT, btn_up)
		r = utility.WriteBit(r, _JOYPAD_LEFT_B_BIT, btn_left)
		r = utility.WriteBit(r, _JOYPAD_RIGHT_A_BIT, btn_right)
	}
	return uint(r)
}

func WriteToMemory(addr uint, value uint) {
	if addr != _JOYPAD_ADDR {
		log.Fatalf("Invalid sound address (0x%8X)", addr)
	}
	// leave bits 6-7 zero and 0-1-2-3 are readonly
	new_actions := utility.GetBit(uint(value), _JOYPAD_ACTIONS_BIT)
	new_directions := utility.GetBit(uint(value), _JOYPAD_DIRECTIONS_BIT)

	if (new_actions == 0x0 && actions == 0x1) || (new_directions == 0x0 && directions == 0x1) {
		interrupts.RequestInterruptJoypad()
	}
	actions = new_actions
	directions = new_directions
}

// ACTIONS
func SetJoypadStart() {
	btn_start = 0
	interrupts.RequestInterruptJoypad()
}
func ClearJoypadStart() {
	btn_start = 1
}
func SetJoypadSelect() {
	btn_select = 0
	interrupts.RequestInterruptJoypad()
}
func ClearJoypadSelect() {
	btn_select = 1
}
func SetJoypadA() {
	btn_a = 0
	interrupts.RequestInterruptJoypad()
}
func ClearJoypadA() {
	btn_a = 1
}
func SetJoypadB() {
	btn_b = 0
	interrupts.RequestInterruptJoypad()
}
func ClearJoypadB() {
	btn_b = 1
}

// DIRECTIONS
func SetJoypadDown() {
	btn_down = 0
	interrupts.RequestInterruptJoypad()
}
func ClearJoypadDown() {
	btn_down = 1
}
func SetJoypadUp() {
	btn_up = 0
	interrupts.RequestInterruptJoypad()
}
func ClearJoypadUp() {
	btn_up = 1
}
func SetJoypadRight() {
	btn_right = 0
	interrupts.RequestInterruptJoypad()
}
func ClearJoypadRight() {
	btn_right = 1
}
func SetJoypadLeft() {
	btn_left = 0
	interrupts.RequestInterruptJoypad()
}
func ClearJoypadLeft() {
	btn_left = 1
}

var ToggleDebugMode func()
var TogglePauseMode func()
var ToggleManualMode func()
var SaveGame func()
