package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	cpu "github.com/giammirove/gampboy_emulator/internal/cpu"
	decoder "github.com/giammirove/gampboy_emulator/internal/decoder"
	"github.com/giammirove/gampboy_emulator/internal/gui"
	"github.com/giammirove/gampboy_emulator/internal/headers"
	"github.com/giammirove/gampboy_emulator/internal/interrupts"
	"github.com/giammirove/gampboy_emulator/internal/joypad"
	mmu "github.com/giammirove/gampboy_emulator/internal/mmu"
	"github.com/giammirove/gampboy_emulator/internal/ppu"
	"github.com/giammirove/gampboy_emulator/internal/serial"
	"github.com/giammirove/gampboy_emulator/internal/sound"
	"github.com/giammirove/gampboy_emulator/internal/timer"
	"github.com/sqweek/dialog"
)

func Init() {
	name := flag.String("r", "", "ROM path (relative)")
	debug := flag.Bool("d", false, "Debug Mode")
	window_debug := flag.Bool("wd", false, "Window debug enabled")
	manual := flag.Bool("m", false, "Manual Mode")
	server := flag.Bool("s", false, "Server Mode")
	scale := flag.Int("sc", 3, "Scale")
	flag.Parse()

	cpu.DEBUG = *debug
	cpu.MANUAL = *manual
	if *name == "" {
		fmt.Printf("!!! ROM not found, opening dialog\n")
		var err error
		*name, err = dialog.File().Title("Choose the ROM").Filter("rom", "gb", "gbc").Load()
		if err != nil {
			fmt.Printf("A ROM path is required !!!\n")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	path := strings.Trim(string(*name), " ")
	rom, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error with ROM\n\t%s", err)

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
	gui.SERVER_MODE = *server
	gui.SCALE = uint(*scale)
	gui.Init()
}

func main() {

	Init()

	go cpu.Run()

	gui.Run()
}
