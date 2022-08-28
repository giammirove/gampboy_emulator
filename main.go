package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	cpu "github.com/giammirove/gbemu/internal/cpu"
	decoder "github.com/giammirove/gbemu/internal/decoder"
	"github.com/giammirove/gbemu/internal/gui"
	"github.com/giammirove/gbemu/internal/headers"
	"github.com/giammirove/gbemu/internal/interrupts"
	"github.com/giammirove/gbemu/internal/joypad"
	mmu "github.com/giammirove/gbemu/internal/mmu"
	"github.com/giammirove/gbemu/internal/ppu"
	"github.com/giammirove/gbemu/internal/serial"
	"github.com/giammirove/gbemu/internal/sound"
	"github.com/giammirove/gbemu/internal/timer"
)

func Init() {
	name := flag.String("r", "", "ROM path (relative)")
	debug := flag.Bool("d", false, "Debug Mode")
	window_debug := flag.Bool("wd", false, "Window debug enabled")
	manual := flag.Bool("m", false, "Manual Mode")
	flag.Parse()

	cpu.DEBUG = *debug
	cpu.MANUAL = *manual
	if *name == "" {
		fmt.Printf("A rom path is required !!!\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	path := strings.Trim(string(*name), " ")
	rom, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error with rom\n\t%s", err)

	}
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

	timer.Cycle = cpu.Cycle
	interrupts.Cycle = cpu.Cycle
	interrupts.MMUWriteToMemory = mmu.WriteToMemory
	interrupts.GetHalted = cpu.GetHalted

	gui.DEBUG_WINDOW = *window_debug
}

func main() {

	Init()

	go cpu.Run()

	gui.Run()
}
