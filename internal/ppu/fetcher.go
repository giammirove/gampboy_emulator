package ppu

import (
	"log"

	"github.com/giammirove/gampboy_emulator/internal/headers"
	"github.com/giammirove/gampboy_emulator/internal/utility"
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
const _FETCHER_STATE_PUSH uint = 3

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

type sprite_t struct {
	addr            uint
	x               uint
	y               uint
	tile_index      uint
	horizontal_flip bool
	vertical_flip   bool
	bg_priority     bool
	palette         bool
	cgb_palette     uint
}

var buffer [_LCD_WIDTH][_LCD_HEIGHT]uint32

// incremented at the last step of fetcher
var fetcher_x uint8

var tile_id uint8
var tile_addr uint
var bg_bank uint
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

var fifo_array [_LCD_WIDTH * _LCD_HEIGHT]uint32
var fifo_len uint
var fifo_front uint
var fifo_back uint

var sprite_pixels [_SPRITES_PER_PIXEL][_PIXELS_LEN]uint

var sprite_tiles []sprite_t

var sprites_on_line [_SPRITES_PER_LINE]uint
var sprites_on_line_len uint
var sprites_map [_PIXELS_LEN]uint

var bg_priority [_LCD_WIDTH]bool
var bg_transparent [_LCD_WIDTH]bool

var state uint
var ticks uint

// incremented on every window pixel
// The window keeps an internal line counter that’s functionally similar to LY,
// and increments alongside it.
// However, it only gets incremented when the window is visible
// This line counter determines what window line is to be rendered on the current scanline.
var window_line_counter uint8

func InitFetcher() {
	fifo_len = 0
	fifo_front = 0
	fifo_back = 0
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

	fifo_len = 0
	fifo_front = 0
	fifo_back = 0

	sprite_tiles = []sprite_t{}
	bg_transparent = [_LCD_WIDTH]bool{}
	bg_priority = [_LCD_WIDTH]bool{}
}
func FetcherOamLoad() {
	sprites_on_line_len = 0
	loadSpritePerLine()
}
func FetcherClearFIFO() {
	fifo_len = 0
	fifo_back = 0
	fifo_front = 0
}

func GetTileAddr() uint {
	y := tile_y
	if tile_addr != 0x0 && GetCGBBGVerticalFlip(tile_addr) {
		y = 8 - y - 1
	}
	return (tiledata_base + uint(tile_id)<<4) + uint(y)<<1
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
		y = s_h - y - 1
	}
	return (_TILE_DATA_AREA_SECONDARY + uint(t_id)<<4) + y<<1
}

func FetcherTick() {

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

	// map_x/8 because a tile is 8x8
	map_x = (fetcher_x + uint8(GetSCX())) / 8
	map_y = (uint8(GetLY()) + uint8(GetSCY())) / 8
	// has no scroll
	win_x = (fetcher_x - (uint8(GetWX()) - 7)) / 8
	// has no scroll
	win_y = window_line_counter / 8
	tile_y = (uint8(GetLY()) + uint8(GetSCY())) % 8

	switch state {
	case _FETCHER_STATE_GET_TILE:
		if IsWindowVisible() && IsPixelInWindow(int(fetcher_x), int(GetLY())) {
			if GetLCDCWinTileMapDisplay() {
				window_tilemap = _TILEMAP_SECONDARY
			} else {
				window_tilemap = _TILEMAP_DEFAULT
			}
			tile_addr = window_tilemap + uint(win_x) + (uint(win_y))<<5
		} else {
			if GetLCDCBGTileMapDisplayArea() {
				bg_tilemap = _TILEMAP_SECONDARY
			} else {
				bg_tilemap = _TILEMAP_DEFAULT
			}
			tile_addr = bg_tilemap + uint(map_x) + uint(map_y)<<5

		}
		// the first tile map is for tile_id
		// the second one for attributes
		tile_id = ReadFromVRAMMemory(tile_addr, 0)
		if !GetLCDCBGWinTileDataArea() {
			// indexing is [-127,+128] , so need to translate to [0,256]
			tile_id += 128
			tiledata_base = _TILE_DATA_AREA_DEFAULT
		} else {
			tiledata_base = _TILE_DATA_AREA_SECONDARY
		}
		if GetLCDCOBJDisplay() {
			sprite_tiles = getSpriteTiles()
			sprite_pixels = [_SPRITES_PER_PIXEL][_PIXELS_LEN]uint{}
		} else {
			sprite_tiles = []sprite_t{}
		}
		fetcher_x += 8
		break
	case _FETCHER_STATE_GET_TILE_DATA0:
		bank := uint(0)
		if headers.IsCGB() && tile_addr != 0x0 && GetCGBBGVRAMBank(tile_addr) {
			bank = 1
		}
		data := ReadFromVRAMMemory(GetTileAddr(), bank)
		for b := _PIXELS_LEN - 1; b >= 0; b-- {
			bg_pixels[b] = utility.GetBit(uint(data), uint(b))
		}
		for i := 0; i < len(sprite_tiles); i++ {
			bank = 0
			if headers.IsCGB() && GetSpriteTileVRAMBankNumber(sprite_tiles[i].addr) {
				bank = 1
			}
			data = ReadFromVRAMMemory(GetSpriteTileAddr(sprite_tiles[i]), bank)
			for b := _PIXELS_LEN - 1; b >= 0; b-- {
				sprite_pixels[i][b] = utility.GetBit(uint(data), uint(b))
			}
		}
		break
	case _FETCHER_STATE_GET_TILE_DATA1:
		bank := uint(0)
		if headers.IsCGB() && tile_addr != 0x0 && GetCGBBGVRAMBank(tile_addr) {
			bank = 1
		}
		data := ReadFromVRAMMemory(GetTileAddr()+1, bank)
		for b := _PIXELS_LEN - 1; b >= 0; b-- {
			bg_pixels[b] |= utility.GetBit(uint(data), uint(b)) << 1
		}
		for i := 0; i < len(sprite_tiles); i++ {
			bank = 0
			if headers.IsCGB() && GetSpriteTileVRAMBankNumber(sprite_tiles[i].addr) {
				bank = 1
			}
			data = ReadFromVRAMMemory(GetSpriteTileAddr(sprite_tiles[i])+1, bank)
			for b := _PIXELS_LEN - 1; b >= 0; b-- {
				sprite_pixels[i][b] |= utility.GetBit(uint(data), uint(b)) << 1
			}
		}
		break
	// case _FETCHER_STATE_IDLE:
	// 	break
	case _FETCHER_STATE_PUSH:
		// no space left
		if fifo_len > 8 {
			return
		}
		// TODO: is this necessary
		x := uint(fetcher_x) - (8 - (GetSCX() % 8))
		if x >= 0 {
			// load fifo
			for i := _PIXELS_LEN - 1; i >= 0; i-- {
				bg_i := i
				if tile_addr != 0x0 && GetCGBBGHorizontalFlip(tile_addr) {
					bg_i = _PIXELS_LEN - 1 - i
				}
				color := GetBGColor(bg_pixels[bg_i])
				if !GetLCDCBGWinDisplay() && headers.IsGB() {
					bg_pixels[bg_i] = 0
					color = GetBGColor(bg_pixels[bg_i])
				}
				if headers.IsCGB() && tile_addr != 0x0 {
					color = GetCGBBGColor(tile_addr, bg_pixels[bg_i])
				}

				if GetLCDCOBJDisplay() && len(sprite_tiles) > 0 {
					bg_priority := false
					if headers.IsCGB() {
						bg_priority = GetLCDCBGWinDisplay() && (tile_addr != 0x0 && GetCGBBGPriority(tile_addr))
					}
					color = fetcherGetSpritePixel(i, bg_pixels[bg_i], color, bg_priority)
				}

				fifo_array[fifo_back] = color
				fifo_len++
				fifo_back++
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

func fetcherGetSpritePixel(index int, bg_color_index uint, bg_color uint32, bg_priority bool) uint32 {
	new_color := bg_color
	new_color_c := uint(0)
	bit := index
	min_x := _LCD_WIDTH + 8
	min_oam := uint(0xFFFF)
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
		if (!bg_priority && !sprite_tiles[s].bg_priority) || IsTransparent(bg_color_index) {
			prior := x < min_x
			// CGB has different priorities
			if headers.IsCGB() {
				prior = sprite_tiles[s].addr < min_oam
			}
			// the `<` avoids the problem of same x
			if prior {
				min_x = x
				min_oam = sprite_tiles[s].addr
				new_color_c = sprite_pixels[s][bit]
				if headers.IsCGB() {
					new_color = GetCGBOBPColor(sprite_tiles[s].addr, new_color_c)
				} else {
					if sprite_tiles[s].palette {
						new_color = GetOBP1Color(new_color_c)
					} else {
						new_color = GetOBP0Color(new_color_c)
					}
				}
			}
		}
	}
	if !IsTransparent(new_color_c) {
		return new_color
	}
	return bg_color
}

func fetcherPixelPush() {
	if fifo_len > _FIFO_MAX_LEN {
		pixel := FetcherFIFOPop()
		if line_x >= uint8(GetSCX())%8 {
			buffer[buffered_x][GetLY()] = pixel
			buffered_x++
		}
		line_x++
	}
}

func loadSpritePerLine() {
	loaded := uint(0)
	// min_x := _LCD_WIDTH + 8
	// min_oam := uint(0xFFFF)
	height := 8
	if GetLCDCOBJSize() {
		height = 16
	}
	for i := uint(0); i < _SPRITES_NUM && loaded < _SPRITES_PER_LINE; i++ {
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
			pal_cgb := GetSpriteCGBPaletteNumber(sprites_on_line[i])
			bg_p := GetSpriteBGtoOAMPriority(sprites_on_line[i])

			tile_addr = append(tile_addr,
				sprite_t{addr: sprites_on_line[i],
					x: s_x, y: s_y,
					tile_index:      t_i,
					horizontal_flip: flip_h,
					vertical_flip:   flip_v,
					palette:         pal,
					cgb_palette:     pal_cgb,
					bg_priority:     bg_p})
			c++
		}

		if c >= _SPRITES_PER_PIXEL {
			break
		}
	}

	return tile_addr
}

func FetcherFIFOPop() uint32 {
	// if fifo.Len() == 0 {
	// 	log.Fatal("fifo is empty")
	// }
	//
	// return fifo.Remove(fifo.Front()).(uint32)
	if fifo_len == 0 {
		log.Fatal("fifo is empty")
	}

	fifo_len--
	r := fifo_array[fifo_front]
	fifo_front++
	return r
}

func DrawScanline() {
	bg_priority = [_LCD_WIDTH]bool{}
	renderBG()
	renderWindow()
	renderSprites()
	buffered_x = _LCD_WIDTH
}
func renderBG() {
	if !GetLCDCBGWinDisplay() && !headers.IsCGB() {
		return
	}
	ly := int(GetLY())
	scx := int(GetSCX())
	scy := int(GetSCY())
	tilemap := _TILEMAP_DEFAULT
	if GetLCDCBGTileMapDisplayArea() {
		tilemap = _TILEMAP_SECONDARY
	}
	if ly >= 144 {
		return
	}
	tiledata := _TILE_DATA_AREA_SECONDARY
	for x := 0; x < _LCD_WIDTH; x += 8 {
		map_x := ((x + scx) % 256) / 8
		map_y := ((ly + scy) % 256) / 8
		tile_y := (ly + scy) % 256 % 8
		tile_addr := tilemap + uint(map_y*32+map_x)
		tile_id = ReadFromVRAMMemory(tile_addr, 0)
		if !GetLCDCBGWinTileDataArea() {
			tile_id += 128
			tiledata = _TILE_DATA_AREA_DEFAULT
		}
		if GetCGBBGVerticalFlip(tile_addr) {
			tile_y = _PIXELS_LEN - 1 - tile_y
		}
		tile_data_addr := uint(tiledata) + uint(tile_id)*16 + uint(tile_y*2)

		bank := uint(0)
		if headers.IsCGB() && GetCGBBGVRAMBank(tile_addr) {
			bank = 1
		}
		byte1 := ReadFromVRAMMemory(tile_data_addr, bank)
		byte2 := ReadFromVRAMMemory(tile_data_addr+1, bank)
		pixels := [_PIXELS_LEN]uint{}
		c := 0
		for b := _PIXELS_LEN - 1; b >= 0; b-- {
			pixels[b] = utility.GetBit(uint(byte1), uint(b))
			pixels[b] |= utility.GetBit(uint(byte2), uint(b)) << 1
			color := GetBGColor(pixels[b])
			if headers.IsCGB() {
				color = GetCGBBGColor(tile_addr, pixels[b])
			}
			pos := x + c
			if GetCGBBGHorizontalFlip(tile_addr) {
				pos = x + (_PIXELS_LEN - 1 - c)
			}
			buffer[pos][ly] = color
			bg_transparent[pos] = IsTransparent(pixels[b])
			if headers.IsCGB() {
				bg_priority[pos] = GetCGBBGPriority(tile_addr) && GetLCDCBGWinDisplay()
			}
			c++
		}
	}
}
func renderWindow() {
	if !IsWindowVisible() {
		return
	}
	ly := int(GetLY())
	wx := int(GetWX())
	wy := int(GetWY())
	// no window
	if ly < wy || wx > _WX_MAX {
		return
	}
	if ly >= 144 {
		return
	}
	wx -= 7
	tilemap := _TILEMAP_DEFAULT
	if GetLCDCWinTileMapDisplay() {
		tilemap = _TILEMAP_SECONDARY
	}
	tiledata := _TILE_DATA_AREA_SECONDARY
	for x := 0; x+wx < _LCD_WIDTH; x += 8 {
		if x+wx < 0 {
			continue
		}
		map_x := x / 8
		map_y := window_line_counter / 8
		tile_y := window_line_counter % 8
		tile_addr := tilemap + uint(uint(map_y)*32+uint(map_x))
		tile_id = ReadFromVRAMMemory(tile_addr, 0)
		if !GetLCDCBGWinTileDataArea() {
			tile_id += 128
			tiledata = _TILE_DATA_AREA_DEFAULT
		}
		if GetCGBBGVerticalFlip(tile_addr) {
			tile_y = _PIXELS_LEN - 1 - tile_y
		}
		tile_data_addr := uint(tiledata) + uint(tile_id)*16 + uint(tile_y*2)

		bank := uint(0)
		if headers.IsCGB() && GetCGBBGVRAMBank(tile_addr) {
			bank = 1
		}
		byte1 := ReadFromVRAMMemory(tile_data_addr, bank)
		byte2 := ReadFromVRAMMemory(tile_data_addr+1, bank)
		pixels := [_PIXELS_LEN]uint{}
		c := 0
		for b := _PIXELS_LEN - 1; b >= 0; b-- {
			pixels[b] = utility.GetBit(uint(byte1), uint(b))
			pixels[b] |= utility.GetBit(uint(byte2), uint(b)) << 1
			color := GetColor(pixels[b])
			if headers.IsCGB() {
				color = GetCGBBGColor(tile_addr, pixels[b])
			}
			pos := x + wx + c
			if GetCGBBGHorizontalFlip(tile_addr) {
				pos = x + wx + (_PIXELS_LEN - 1 - c)
			}
			if pos < _LCD_WIDTH {
				buffer[pos][ly] = color
				bg_transparent[pos] = IsTransparent(pixels[b])
				if headers.IsCGB() {
					bg_priority[pos] = GetCGBBGPriority(tile_addr) && GetLCDCBGWinDisplay()
				}
				c++
			} else {
				return
			}
		}
	}
}
func renderSprites() {
	if !GetLCDCOBJDisplay() {
		return
	}
	loaded := uint(0)
	ly := int(GetLY())
	if ly >= 144 {
		return
	}
	height := 8
	if GetLCDCOBJSize() {
		height = 16
	}
	row := [_LCD_WIDTH]int{}
	for x := 0; x < _LCD_WIDTH; x++ {
		row[x] = _LCD_WIDTH + 1
	}
	for i := uint(0); i < _SPRITES_NUM; i++ {
		addr := _OAM_SPRITES_BASE + i*4
		x := int(GetSpriteXPosition(addr)) - 8
		y := int(GetSpriteYPosition(addr)) - 16
		if y > ly || y+height <= ly {
			continue
		}
		tile_index := uint(GetSpriteTileIndex(addr))
		if GetLCDCOBJSize() {
			tile_index = utility.ClearBit(tile_index, 0)
		}
		y = (ly - y)
		if GetSpriteVerticalFlip(addr) {
			y = height - y - 1
		}
		spritedata := (_TILE_DATA_AREA_SECONDARY + uint(tile_index)<<4) + uint(y)<<1

		bank := uint(0)
		if headers.IsCGB() && GetSpriteTileVRAMBankNumber(addr) {
			bank = 1
		}
		byte1 := ReadFromVRAMMemory(spritedata, bank)
		byte2 := ReadFromVRAMMemory(spritedata+1, bank)

		pixels := [_PIXELS_LEN]uint{}
		c := 0
		for b := _PIXELS_LEN - 1; b >= 0; b-- {
			pos := x + c
			if GetSpriteHorizontalFlip(addr) {
				pos = x + (_PIXELS_LEN - 1 - c)
			}
			c++
			if pos < 0 || pos >= _LCD_WIDTH {
				continue
			}
			to_display := x < row[pos]
			// CGB has different priorities
			if headers.IsCGB() {
				to_display = row[pos] == _LCD_WIDTH+1
			}
			if !to_display {
				continue
			}
			pixels[b] = utility.GetBit(uint(byte1), uint(b))
			pixels[b] |= utility.GetBit(uint(byte2), uint(b)) << 1
			if IsTransparent(pixels[b]) {
				continue
			}
			color := GetColor(0)
			if headers.IsGB() {
				if GetSpritePaletteNumber(addr) {
					color = GetOBP1Color(pixels[b])
				} else {
					color = GetOBP0Color(pixels[b])
				}
			}
			if headers.IsCGB() {
				color = GetCGBOBPColor(addr, pixels[b])
			}
			if (!bg_priority[pos] && !GetSpriteBGtoOAMPriority(addr)) || bg_transparent[pos] {
				buffer[pos][ly] = color
				row[pos] = x
			}
		}
		loaded++
		if loaded >= _SPRITES_PER_LINE {
			break
		}
	}
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
		p1 := ReadFromVRAMMemory(addr+c, 0)
		c++
		p2 := ReadFromVRAMMemory(addr+c, 0)
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
