package ppu

import (
	"container/list"
	"github.com/giammirove/gbemu/libs/utility"
	"log"
)

// thanks to
// https://blog.tigris.fr/2019/09/15/writing-an-emulator-the-first-pixel/

const _TILEMAP_DEFAULT uint = 0x9800
const _TILEMAP_SECONDARY uint = 0x9C00
const _TILE_DATA_AREA_DEFAULT uint = 0x8800
const _TILE_DATA_AREA_SECONDARY uint = 0x8000

// Fetcher states
const _FETCHER_STATE_GET_TILE uint = 0
const _FETCHER_STATE_GET_TILE_DATA0 uint = 1
const _FETCHER_STATE_GET_TILE_DATA1 uint = 2
const _FETCHER_STATE_IDLE uint = 3
const _FETCHER_STATE_PUSH uint = 4

const _FIFO_MAX_LEN = 8
const _PIXELS_LEN = 8

const _OAM_SPRITES_BASE uint = 0xFE00
const _SPRITES_NUM uint = 40
const _SPRITES_PER_LINE uint = 10
const _SPRITES_PER_PIXEL uint = 3
const _NULL_SPRITE uint = 0x0

const _LCD_WIDTH = 160
const _LCD_HEIGHT = 144
const _WX_MAX = _LCD_WIDTH + 6
const _WY_MAX = _LCD_HEIGHT - 1

type pixel_t struct {
	x        uint
	y        uint
	color    uint32
	tile_id  uint8
	priority bool
}

type sprite_t struct {
	addr            uint
	x               uint
	y               uint
	tile_index      uint
	horizontal_flip bool
	vertical_flip   bool
	bg_priority     bool
	palette         bool
}

var buffer [_LCD_WIDTH][_LCD_HEIGHT]uint32

// incremented at the last step of fetcher
var fetcher_x uint8

var tile_id uint8
var bg_tilemap uint
var window_tilemap uint
var tiledata_base uint

var map_x uint8
var map_y uint8
var win_x uint8
var win_y uint8
var tile_x uint8
var tile_y uint8
var line_x uint8

// higher pixel x that has been rendered
var buffered_x uint8

// higher pixel x that has been pushed to fifo
var fifo_x uint8

var bg_pixels [_PIXELS_LEN]uint
var fifo *list.List

var sprite_pixels [_SPRITES_PER_PIXEL][_PIXELS_LEN]uint

var sprite_tiles []sprite_t

var sprites_on_line [_SPRITES_PER_LINE]uint
var sprites_on_line_len uint

var state uint
var ticks uint

// incremented on every window pixel
// The window keeps an internal line counter that’s functionally similar to LY,
// and increments alongside it.
// However, it only gets incremented when the window is visible
// This line counter determines what window line is to be rendered on the current scanline.
var window_line_counter uint8

func InitFetcher() {
	fifo = list.New()
	window_line_counter = 0
	line_x = 0
	fetcher_x = 0
	buffered_x = 0
	fifo_x = 0
	tile_id = 0
	tiledata_base = _TILE_DATA_AREA_DEFAULT
	bg_tilemap = _TILEMAP_SECONDARY
	window_tilemap = _TILEMAP_SECONDARY
}

func FetcherStart() {
	state = _FETCHER_STATE_GET_TILE

	line_x = 0
	fetcher_x = 0
	buffered_x = 0
	fifo_x = 0
	tile_id = 0
	tiledata_base = _TILE_DATA_AREA_DEFAULT
	bg_tilemap = _TILEMAP_SECONDARY
	window_tilemap = _TILEMAP_SECONDARY
	fifo = list.New()

	sprite_tiles = []sprite_t{}
}
func FetcherOamLoad() {
	sprites_on_line_len = 0
	loadSpritePerLine()
}
func FetcherClearFIFO() {
	fifo = list.New()
}

func GetTileAddr() uint {
	return (tiledata_base + uint(tile_id)<<4) + uint(tile_y)<<1
}
func GetSpriteTileAddr(sprite sprite_t) uint {
	t_id := sprite.tile_index
	s_h := uint(8)
	if GetLCDCOBJSize() {
		t_id = utility.ClearBit(t_id, 0)
		s_h = 16
	}
	y := (GetLY() - (sprite.y - 16))
	if sprite.vertical_flip {
		// TODO: check why this -1
		y = s_h - y - 1
	}
	return (_TILE_DATA_AREA_SECONDARY + uint(t_id)<<4) + y<<1
}

func FetcherTick() {

	// map_x/8 because a tile is 8x8
	map_x = (fetcher_x + uint8(GetSCX())) / 8
	map_y = (uint8(GetLY()) + uint8(GetSCY())) / 8
	// has no scroll
	win_x = (fetcher_x - (uint8(GetWX()) - 7)) / 8
	// has no scroll
	win_y = window_line_counter / 8
	tile_y = (uint8(GetLY()) + uint8(GetSCY())) % 8

	fetcherCycle()
	// done every tick
	fetcherPixelPush()
}

func fetcherCycle() {

	ticks++
	// fetcher goes at half speed
	// in other words all phases take 2 dots in LCDTick()
	if ticks < 2 {
		return
	}
	ticks = 0

	switch state {
	case _FETCHER_STATE_GET_TILE:
		if GetLCDCBGWinDisplay() {
			addr := uint(0x0)
			if IsWindowVisible() && IsPixelInWindow(int(fetcher_x), int(GetLY())) {
				if GetLCDCWinTileMapDisplay() {
					window_tilemap = _TILEMAP_SECONDARY
				} else {
					window_tilemap = _TILEMAP_DEFAULT
				}
				addr = window_tilemap + uint(win_x) + (uint(win_y))<<5
			} else {
				if GetLCDCBGTileMapDisplayArea() {
					bg_tilemap = _TILEMAP_SECONDARY
				} else {
					bg_tilemap = _TILEMAP_DEFAULT
				}
				addr = bg_tilemap + uint(map_x) + uint(map_y)<<5

			}
			tile_id = ReadFromVRAMMemory(addr)
			if !GetLCDCBGWinTileDataArea() {
				// indexing is [-127,+128] , so need to translate to [0,256]
				tile_id += 128
				tiledata_base = _TILE_DATA_AREA_DEFAULT
			} else {
				tiledata_base = _TILE_DATA_AREA_SECONDARY
			}
		}
		if GetLCDCOBJDisplay() {
			sprite_tiles = getSpriteTiles()
			sprite_pixels = [_SPRITES_PER_PIXEL][_PIXELS_LEN]uint{}
		}
		fetcher_x += 8
		break
	case _FETCHER_STATE_GET_TILE_DATA0:
		data := ReadFromVRAMMemory(GetTileAddr())
		for b := _PIXELS_LEN - 1; b >= 0; b-- {
			bg_pixels[b] = utility.GetBit(uint(data), uint(b))
		}
		for i := 0; i < len(sprite_tiles); i++ {
			data = ReadFromVRAMMemory(GetSpriteTileAddr(sprite_tiles[i]))
			for b := _PIXELS_LEN - 1; b >= 0; b-- {
				sprite_pixels[i][b] = utility.GetBit(uint(data), uint(b))
			}
		}
		break
	case _FETCHER_STATE_GET_TILE_DATA1:
		data := ReadFromVRAMMemory(GetTileAddr() + 1)
		for b := _PIXELS_LEN - 1; b >= 0; b-- {
			bg_pixels[b] |= utility.GetBit(uint(data), uint(b)) << 1
		}
		for i := 0; i < len(sprite_tiles); i++ {
			data = ReadFromVRAMMemory(GetSpriteTileAddr(sprite_tiles[i]) + 1)
			for b := _PIXELS_LEN - 1; b >= 0; b-- {
				sprite_pixels[i][b] |= utility.GetBit(uint(data), uint(b)) << 1
			}
		}
		break
	case _FETCHER_STATE_IDLE:
		break
	case _FETCHER_STATE_PUSH:
		// no space left
		if fifo.Len() > 8 {
			return
		}
		// TODO: is this necessary
		x := uint(fetcher_x) - (8 - (GetSCX() % 8))
		if x >= 0 {
			// load fifo
			for i := _PIXELS_LEN - 1; i >= 0; i-- {
				color := GetBGColor(bg_pixels[i])
				// x_draw := fetcher_x
				if !GetLCDCBGWinDisplay() {
					color = GetBGColor(0)
					bg_pixels[i] = 0
				}

				if GetLCDCOBJDisplay() && len(sprite_tiles) > 0 {
					color = fetcherGetSpritePixel(i, bg_pixels[i])
				}

				p := pixel_t{x: uint(fetcher_x), y: GetLY(), color: color, tile_id: tile_id}
				fifo.PushBack(p)
				fifo_x++
			}
		} else {
			log.Printf("x nop")
		}
		break
	default:
		log.Fatalf("Fetcher state not recognized %d\n", state)
		break
	}

	state++
	if state > _FETCHER_STATE_PUSH {
		state = _FETCHER_STATE_GET_TILE
	}

}

func fetcherGetSpritePixel(index int, bg_color uint) uint32 {
	original := GetBGColor(bg_color)
	new_color := original
	new_color_c := uint(0)
	bit := index
	min_x := _LCD_WIDTH + 8
	for s := 0; s < len(sprite_tiles); s++ {
		x := int(sprite_tiles[s].x) - 8 + int(GetSCX()%8)
		if x+8 < int(fifo_x) {
			continue
		}
		bit = int(fifo_x) - int(x)
		// check if bit is already displayed or if out of bounds
		if bit < 0 || bit > 7 {
			continue
		}
		if !sprite_tiles[s].horizontal_flip {
			bit = 7 - bit
		}
		// do this check before update x is important
		if IsTransparent(sprite_pixels[s][bit]) {
			continue
		}
		if !sprite_tiles[s].bg_priority || IsTransparent(bg_color) {
			// the `<` avoids the problem of same x
			if x < min_x {
				min_x = x
				if sprite_tiles[s].palette {
					new_color = GetOBP1Color(sprite_pixels[s][bit])
				} else {
					new_color = GetOBP0Color(sprite_pixels[s][bit])
				}
				new_color_c = sprite_pixels[s][bit]
			}
		}
	}
	if !IsTransparent(new_color_c) {
		return new_color
	}
	return original
}

func fetcherPixelPush() {
	if fifo.Len() > _FIFO_MAX_LEN {
		pixel := FetcherFIFOPop()
		if line_x >= uint8(GetSCX())%8 {
			buffer[buffered_x][pixel.y] = pixel.color
			buffered_x++
		}
		line_x++
	}
}

func loadSpritePerLine() {
	loaded := uint(0)
	height := 8
	if GetLCDCOBJSize() {
		height = 16
	}
	for i := uint(0); i < _SPRITES_NUM; i++ {
		addr := _OAM_SPRITES_BASE + i*4
		x := int(GetSpriteXPosition(addr))
		// not visible
		if x == 0 {
			continue
		}
		y := int(GetSpriteYPosition(addr)) - 16
		if y <= int(GetLY()) && y+height > int(GetLY()) {
			sprites_on_line[loaded] = addr
			loaded++
		}
		// check if too much
		if loaded >= _SPRITES_PER_LINE {
			break
		}
	}
	sprites_on_line_len = uint(loaded)
}

func getSpriteTiles() []sprite_t {
	tile_addr := []sprite_t{}
	c := uint(0)
	for i := 0; i < int(sprites_on_line_len); i++ {
		// int because x could be also -7 (one pixel will be displayed)
		x := int(GetSpriteXPosition(sprites_on_line[i])) - 8 + int(GetSCX())%8
		// check that sprite is in this row of 8 pixel
		// choose the one with lower x
		// < and not <= because in case of parity of lower x we have to choose
		// the first one in oam
		// x+8 > fetcher_x ( > is strict !! important)
		if (x < int(fetcher_x) && x+8 >= int(fetcher_x)) || (x >= int(fetcher_x) && x < int(fetcher_x)+8) {
			s_x := GetSpriteXPosition(sprites_on_line[i])
			s_y := GetSpriteYPosition(sprites_on_line[i])
			t_i := GetSpriteTileIndex(sprites_on_line[i])
			flip_h := GetSpriteHorizontalFlip(sprites_on_line[i])
			flip_v := GetSpriteVerticalFlip(sprites_on_line[i])
			pal := GetSpritePaletteNumber(sprites_on_line[i])
			bg_p := GetSpriteBGtoOAMPriority(sprites_on_line[i])

			tile_addr = append(tile_addr,
				sprite_t{addr: sprites_on_line[i],
					x: s_x, y: s_y,
					tile_index:      t_i,
					horizontal_flip: flip_h,
					vertical_flip:   flip_v,
					palette:         pal,
					bg_priority:     bg_p})
			c++
		}

		if c >= _SPRITES_PER_PIXEL {
			break
		}
	}

	return tile_addr
}

func FetcherFIFOPop() pixel_t {
	if fifo.Len() == 0 {
		log.Fatal("fifo is empty")
	}

	return fifo.Remove(fifo.Front()).(pixel_t)
}
func FetcherGetCurrentDots() uint {
	return current_dots
}
func FetcherGetPushedX() uint {
	return uint(buffered_x)
}
func FetcherGetY() uint {
	return GetLY()
}
func FetcherGetBuffer() [_LCD_WIDTH][_LCD_HEIGHT]uint32 {
	return buffer
}
func IsWindowVisible() bool {
	return GetLCDCWinDisplay() && (int(GetWX()) >= 0 && int(GetWX()) <= _WX_MAX && int(GetWY()) >= 0 && int(GetWY()) <= _WY_MAX)
}
func IsPixelInWindow(x int, y int) bool {
	return (x >= int(GetWX())-7 && x < int(GetWX())-7+_LCD_WIDTH) && (y >= int(GetWY()) && y < int(GetWY()+_LCD_HEIGHT))
}
func IncrementWindowLineCounter() {
	// if is in bounds
	if IsWindowVisible() && GetWY() < GetLY() && GetWY()+_LCD_HEIGHT > GetLY() {
		window_line_counter++
	}
}
func ResetWindowLineCounter() {
	window_line_counter = 0
}

// this will return a tile 8x8 pixel as matrix
func GetTileData(addr uint) [_TILE_W][_TILE_H]uint {
	tile := [_TILE_W][_TILE_H]uint{}
	c := uint(0)
	y := uint(0)
	for t := 0; t < _TILE_BYTES; t += 2 {
		p1 := ReadFromVRAMMemory(addr + c)
		c++
		p2 := ReadFromVRAMMemory(addr + c)
		c++
		x := 0
		for b := 7; b >= 0; b-- {
			tile[x][y] = utility.GetBit(uint(p1), uint(b)) | (utility.GetBit(uint(p2), uint(b)) << 1)
			x++
		}
		y++
	}
	return tile
}
