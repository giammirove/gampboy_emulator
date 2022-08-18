package gui

import (
	"log"

	"github.com/giammirove/gbemu/libs/joypad"
	"github.com/giammirove/gbemu/libs/mmu"
	"github.com/giammirove/gbemu/libs/ppu"
	"github.com/veandco/go-sdl2/sdl"
)

const SCREEN_W = 256
const SCREEN_H = 256
const DEBUG_W = 144
const DEBUG_H = 216
const WIDTH = uint(160)
const HEIGHT = uint(144)

// var context *cairo.Context
// var is_context bool

var sdl_surface *sdl.Surface
var sdl_window *sdl.Window

var sdl_surface2 *sdl.Surface
var sdl_window2 *sdl.Window

const _DEBUG_WINDOW bool = false

var scale = uint(4)

func Run() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	var err error
	sdl_window, err = sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(WIDTH*scale), int32(HEIGHT*scale), sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer sdl_window.Destroy()

	sdl_surface, err = sdl_window.GetSurface()
	if err != nil {
		panic(err)
	}
	sdl_surface.FillRect(nil, 0)

	if _DEBUG_WINDOW {
		sdl_window2, err = sdl.CreateWindow("test2", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, int32(DEBUG_W*scale), int32(DEBUG_H*scale), sdl.WINDOW_SHOWN)
		sdl_window2.SetPosition(0, 0)
		if err != nil {
			panic(err)
		}
		defer sdl_window2.Destroy()

		sdl_surface2, err = sdl_window2.GetSurface()
		if err != nil {
			panic(err)
		}
		sdl_surface2.FillRect(nil, 0)
	}

	sdl_window2.UpdateSurface()
	sdl_window.UpdateSurface()

	running := true
	var prev_frame = 0
	for running {
		if prev_frame != ppu.GetCurrentFrame() {
			prev_frame = ppu.GetCurrentFrame()
			if _DEBUG_WINDOW {
				UpdateGUI()
			}
			UpdateGUI3()
		}
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			case *sdl.KeyboardEvent:
				ev := event.(*sdl.KeyboardEvent)
				switch ev.Keysym.Scancode {
				case sdl.SCANCODE_RETURN:
					if ev.Type == sdl.KEYDOWN {
						joypad.SetJoypadStart()
					} else {
						joypad.ClearJoypadStart()
					}
					break
				case sdl.SCANCODE_SPACE:
					if ev.Type == sdl.KEYDOWN {
						joypad.SetJoypadSelect()
					} else {
						joypad.ClearJoypadSelect()
					}
					break
				case sdl.SCANCODE_Z:
					if ev.Type == sdl.KEYDOWN {
						joypad.SetJoypadA()
					} else {
						joypad.ClearJoypadA()
					}
					break
				case sdl.SCANCODE_X:
					if ev.Type == sdl.KEYDOWN {
						joypad.SetJoypadB()
					} else {
						joypad.ClearJoypadB()
					}
					break
				case sdl.SCANCODE_DOWN:
					if ev.Type == sdl.KEYDOWN {
						joypad.SetJoypadDown()
					} else {
						joypad.ClearJoypadDown()
					}
					break
				case sdl.SCANCODE_UP:
					if ev.Type == sdl.KEYDOWN {
						joypad.SetJoypadUp()
					} else {
						joypad.ClearJoypadUp()
					}
					break
				case sdl.SCANCODE_RIGHT:
					if ev.Type == sdl.KEYDOWN {
						joypad.SetJoypadRight()
					} else {
						joypad.ClearJoypadRight()
					}
					break
				case sdl.SCANCODE_LEFT:
					if ev.Type == sdl.KEYDOWN {
						joypad.SetJoypadLeft()
					} else {
						joypad.ClearJoypadLeft()
					}
					break
				case sdl.SCANCODE_D:
					if ev.Type == sdl.KEYDOWN {
						joypad.ToggleDebugMode()
					}
				case sdl.SCANCODE_P:
					if ev.Type == sdl.KEYDOWN {
						joypad.TogglePauseMode()
					}
				case sdl.SCANCODE_M:
					if ev.Type == sdl.KEYDOWN {
						joypad.ToggleManualMode()
					}
				case sdl.SCANCODE_S:
					if ev.Type == sdl.KEYDOWN {
						joypad.SaveGame()
					}
				}
				break
			}
		}
	}
}

func ColorPixel(x uint, y uint, color uint32) {
	rect := sdl.Rect{X: int32(x), Y: int32(y), W: int32(scale), H: int32(scale)}
	sdl_surface.FillRect(&rect, color)
}
func ColorPixel2(x uint, y uint, color uint32) {
	rect := sdl.Rect{X: int32(x), Y: int32(y), W: int32(scale), H: int32(scale)}
	sdl_surface2.FillRect(&rect, color)
}

func RefreshGUI() {
	sdl_window.UpdateSurface()
}
func RefreshGUI2() {
	sdl_window2.UpdateSurface()
}

var _x uint
var _y uint

func ShowTile(base uint, base_x uint, base_y uint) {
	tile := ppu.GetTileData(base)
	for x := uint(0); x < uint(len(tile)); x++ {
		for y := uint(0); y < uint(len(tile[x])); y++ {
			ColorPixel2(base_x+x*scale, base_y+y*scale, ppu.GetBGColor(tile[x][y]))
		}
	}
}

func UpdateGUI() {
	const addr = 0x8000
	y_d := uint(0)
	x_d := uint(0)
	tile_n := uint(0)
	for _y = uint(0); _y < 24; _y++ {
		// ShowTile(addr+(y*16), x_d, y_d)
		for _x = uint(0); _x < 16; _x++ {
			ShowTile(addr+tile_n*16, x_d+_x*scale, y_d+_y*scale)
			if addr+tile_n*16 == 0x800F {
				log.Fatalf("%02X", tile_n)
			}
			x_d += 8 * scale
			tile_n++
		}
		x_d = 0
		y_d += 8 * scale
	}
	RefreshGUI2()
}
func UpdateGUI2() {
	const addr = 0x8000
	y_d := uint(0)
	x_d := uint(0)
	tile_n := uint(0)
	tile_index := uint(0)
	for _y := uint(0); _y < 32; _y++ {
		// ShowTile(addr+(y*16), x_d, y_d)
		for _x = uint(0); _x < 32; _x++ {
			tile_n = mmu.ReadFromMemory(0x9800 + tile_index)
			ShowTile(addr+tile_n*16, x_d+_x*scale, y_d+_y*scale)
			if _y == 4 && _x == 3 {
				// log.Printf("!! tile id %d | addr %04X\n", tile_n, 0x9800+tile_index)
			}
			x_d += 8 * scale
			tile_index++
		}
		x_d = 0
		y_d += 8 * scale
	}
	RefreshGUI2()
}

func UpdateGUI3() {
	buffer := ppu.FetcherGetBuffer()
	for x := uint(0); x < 160; x++ {
		for y := uint(0); y < 144; y++ {
			ColorPixel(x*scale, y*scale, buffer[x][y])
		}
	}
	RefreshGUI()
}

func DelayGUI(delay uint32) {
	sdl.Delay(delay)
}
func TicksGUI() uint32 {
	return sdl.GetTicks()
}
