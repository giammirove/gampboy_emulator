package main

import (
	"flag"
	"io/ioutil"
	"log"

	cpu "github.com/giammirove/gbemu/libs/cpu"
	decoder "github.com/giammirove/gbemu/libs/decoder"
	"github.com/giammirove/gbemu/libs/gui"
	"github.com/giammirove/gbemu/libs/headers"
	"github.com/giammirove/gbemu/libs/interrupts"
	"github.com/giammirove/gbemu/libs/joypad"
	mmu "github.com/giammirove/gbemu/libs/mmu"
	"github.com/giammirove/gbemu/libs/ppu"
	"github.com/giammirove/gbemu/libs/serial"
	"github.com/giammirove/gbemu/libs/sound"
	"github.com/giammirove/gbemu/libs/timer"
)

func Init() {
	name := flag.String("r", "", "ROM path inside ./roms/")
	debug := flag.Bool("d", false, "Debug Mode")
	manual := flag.Bool("m", false, "Manual Mode")
	flag.Parse()

	cpu.DEBUG = *debug
	cpu.MANUAL = *manual

	path := "./roms/" + *name
	log.Printf(path)
	rom, _ := ioutil.ReadFile(path)
	headers.Init(rom)
	timer.Init()
	mmu.InitMMU(rom, path)
	decoder.InitDecoder()
	cpu.InitCPU()
	interrupts.Init()
	interrupts.StackPush = cpu.StackPush
	sound.Init()
	serial.Init()
	ppu.ReadFromMemory = mmu.ReadFromMemory
	ppu.Init()
	decoder.Cycle = cpu.Cycle
	joypad.Init()
	joypad.ToggleDebugMode = cpu.ToggleDebugMode
	joypad.TogglePauseMode = cpu.TogglePauseMode
	joypad.ToggleManualMode = cpu.ToggleManualMode
	joypad.SaveGame = mmu.SaveMemory

	ppu.DelayGUI = gui.DelayGUI
	ppu.TicksGUI = gui.TicksGUI
	ppu.ColorPixel = gui.ColorPixel
	ppu.RefreshGUI = gui.RefreshGUI

	timer.Cycle = cpu.Cycle
	interrupts.Cycle = cpu.Cycle
	interrupts.MMUWriteToMemory = mmu.WriteToMemory
	interrupts.GetHalted = cpu.GetHalted
}

func main() {

	Init()

	go cpu.Run()

	gui.Run()
}
