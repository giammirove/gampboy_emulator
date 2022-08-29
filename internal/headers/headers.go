package headers

import (
	"fmt"
	"log"

	"github.com/giammirove/gampboy_emulator/internal/utility"
)

const _HEADER_START = 0x0100
const _HEADER_END = 0x014F
const _HEADER_SIZE = (_HEADER_END - _HEADER_START) + 1
const _BYTE_SIZE = 64

type header_meta_t struct {
	start uint
	end   uint
}

type headers_meta_t struct {
	entry_point             header_meta_t
	nintendo_logo           header_meta_t
	title                   header_meta_t
	manufacturer_code       header_meta_t
	cgb_flag                header_meta_t
	new_licensee_code       header_meta_t
	sgb_flag                header_meta_t
	cartridge_type          header_meta_t
	rom_size                header_meta_t
	ram_size                header_meta_t
	destination_code        header_meta_t
	old_licensee_code       header_meta_t
	mask_rom_version_number header_meta_t
	header_checksum         header_meta_t
	global_checksum         header_meta_t
}

type headers_t struct {
	entry_point             [4]uint8
	nintendo_logo           [0x30]uint8
	title                   string
	manufacturer_code       uint16 //; part of title
	cgb_flag                uint8  //; part of tilte
	new_licensee_code       uint16
	sgb_flag                uint8
	cartridge_type          uint8
	rom_size                uint8
	ram_size                uint8
	destination_code        uint8
	old_licensee_code       uint8
	mask_rom_version_number uint8
	header_checksum         uint8
	global_checksum         uint16
}

var licensee_code_map map[uint16]string = map[uint16]string{
	0x00: "None",
	0x01: "Nintendo R&D1",
	0x08: "Capcom",
	0x13: "Electronic Arts",
	0x18: "Hudson Soft",
	0x19: "b-ai",
	0x20: "kss",
	0x22: "pow",
	0x24: "PCM Complete",
	0x25: "san-x",
	0x28: "Kemco Japan",
	0x29: "seta",
	0x30: "Viacom",
	0x31: "Nintendo",
	0x32: "Bandai",
	0x33: "Ocean/Acclaim",
	/* 34	Konami
	   35	Hector
	   37	Taito
	   38	Hudson
	   39	Banpresto
	   41	Ubi Soft
	   42	Atlus
	   44	Malibu
	   46	angel
	   47	Bullet-Proof
	   49	irem
	   50	Absolute
	   51	Acclaim
	   52	Activision
	   53	American sammy
	     54	Konami
	   55	Hi tech entertainment
	   56	LJN
	   57	Matchbox
	   58	Mattel
	   59	Milton Bradley
	   60	Titus
	   61	Virgin
	   64	LucasArts
	   67	Ocean
	   69	Electronic Arts
	   70	Infogrames
	   71	Interplay
	   72	Broderbund
	   73	sculptured
	   75	sci
	   78	THQ
	   79	Accolade
	   80	misawa
	   83	lozc
	   86	Tokuma Shoten Intermedia
	   87	Tsukuda Original
	   91	Chunsoft
	   92	Video system
	   93	Ocean/Acclaim
	   95	Varie
	   96	Yonezawa/s’pal
	   97	Kaneko
	   99	Pack in soft
	   A4	Konami (Yu-Gi-Oh!)
	*/
}
var cartbridge_type_map map[uint8]string = map[uint8]string{
	0x00: "ROM ONLY",
	0x01: "MBC1",
	0x02: "MBC1+RAM",
	0x03: "MBC1+RAM+BATTERY",
	0x05: "MBC2",
	0x06: "MBC2+BATTERY",
	0x08: "ROM+RAM 1",
	0x09: "ROM+RAM+BATTERY 1",
	0x0B: "MMM01",
	0x0C: "MMM01+RAM",
	0x0D: "MMM01+RAM+BATTERY",
	0x0F: "MBC3+TIMER+BATTERY",
	0x10: "MBC3+TIMER+RAM+BATTERY 2",
	0x11: "MBC3",
	0x12: "MBC3+RAM 2",
	0x13: "MBC3+RAM+BATTERY 2",
	0x19: "MBC5",
	0x1A: "MBC5+RAM",
	0x1B: "MBC5+RAM+BATTERY",
	0x1C: "MBC5+RUMBLE",
	0x1D: "MBC5+RUMBLE+RAM",
	0x1E: "MBC5+RUMBLE+RAM+BATTERY",
	0x20: "MBC6",
	0x22: "MBC7+SENSOR+RUMBLE+RAM+BATTERY",
	0xFC: "POCKET CAMERA",
	0xFD: "BANDAI TAMA5",
	0xFE: "HuC3",
	0xFF: "HuC1+RAM+BATTERY",
}
var cartbridge_with_battery []uint = []uint{0x3, 0x6, 0x9, 0xD, 0xF, 0x10, 0x13, 0x1B, 0x1E, 0x22, 0xFF}

type rom_t struct {
	size  uint
	banks uint
	bits  uint
}

var rom_size_map map[uint8]rom_t = map[uint8]rom_t{
	0x00: {size: 32, banks: 2, bits: 1},
	0x01: {size: 64, banks: 4, bits: 2},
	0x02: {size: 128, banks: 8, bits: 3},
	0x03: {size: 256, banks: 16, bits: 4},
	0x04: {size: 512, banks: 32, bits: 5},
	0x05: {size: 1024, banks: 64, bits: 6},
	0x06: {size: 2048, banks: 128, bits: 7},
	0x07: {size: 4096, banks: 256, bits: 8},
	0x08: {size: 8192, banks: 512, bits: 9},
	0x52: {size: 1127, banks: 72, bits: 0},
	0x53: {size: 1229, banks: 80, bits: 0},
	0x54: {size: 1536, banks: 96, bits: 0},
}
var ram_size_map map[uint8]uint = map[uint8]uint{
	0x0: 8,
	0x1: 8,
	0x2: 8,
	0x3: 32,
	0x4: 128,
	0x5: 64,
}
var destination_code_map map[uint8]string = map[uint8]string{
	0x0: "Japan (and possibly oveeseas)",
	0x1: "Overseas only",
}

func create_header(start uint, end uint) header_meta_t {
	header := header_meta_t{start: start, end: end}
	return header
}

func create_headers_struct() headers_meta_t {
	headers := headers_meta_t{}
	headers.entry_point = create_header(0x0100, 0x0103)
	headers.nintendo_logo = create_header(0x0104, 0x0133)
	headers.title = create_header(0x0134, 0x0143)
	headers.manufacturer_code = create_header(0x013F, 0x0142)
	headers.cgb_flag = create_header(0x0143, 0x0143)
	headers.new_licensee_code = create_header(0x0144, 0x0145)
	headers.sgb_flag = create_header(0x0146, 0x0146)
	headers.cartridge_type = create_header(0x0147, 0x0147)
	headers.rom_size = create_header(0x0148, 0x0148)
	headers.ram_size = create_header(0x0149, 0x0149)
	headers.destination_code = create_header(0x014A, 0x014A)
	headers.old_licensee_code = create_header(0x014B, 0x014B)
	headers.mask_rom_version_number = create_header(0x014C, 0x014C)
	headers.header_checksum = create_header(0x014D, 0x014D)
	headers.global_checksum = create_header(0x014E, 0x014F)

	return headers
}

var headers_meta = create_headers_struct()
var headers headers_t

func getHeaderFromRaw(raw []byte, h header_meta_t) []byte {
	return raw[h.start : h.end+1]
}

func Init(raw []byte) {

	headers = headers_t{}
	headers.title = string(getHeaderFromRaw(raw, headers_meta.title))
	headers.cgb_flag = getHeaderFromRaw(raw, headers_meta.cgb_flag)[0]

	licensee_code := getHeaderFromRaw(raw, headers_meta.new_licensee_code)
	headers.new_licensee_code = uint16(licensee_code[1])

	headers.sgb_flag = getHeaderFromRaw(raw, headers_meta.sgb_flag)[0]
	headers.cartridge_type = getHeaderFromRaw(raw, headers_meta.cartridge_type)[0]
	headers.rom_size = getHeaderFromRaw(raw, headers_meta.rom_size)[0]
	headers.ram_size = getHeaderFromRaw(raw, headers_meta.ram_size)[0]
	headers.destination_code = getHeaderFromRaw(raw, headers_meta.destination_code)[0]

	headers.header_checksum = getHeaderFromRaw(raw, headers_meta.header_checksum)[0]

	// check checksum
	check := uint8(0)
	for a := 0x134; a <= 0x14C; a++ {
		check = check - raw[a] - 1
	}
	if check != headers.header_checksum {
		log.Fatal("Checksum failed!")
	}

	fmt.Printf("%-18s: %s\n", "TITLE", headers.title)
	if headers.cgb_flag == 0xC0 {
		fmt.Printf("%-18s: CGB (%02X)\n", "GB TYPE", headers.cgb_flag)
	} else if headers.cgb_flag == 0x80 {
		fmt.Printf("%-18s: CGB but backwards compatible with NON-CGB (%02X)\n", "GB TYPE", headers.cgb_flag)
	} else {
		fmt.Printf("%-18s: NON-CGB (%02X)\n", "GB TYPE", headers.cgb_flag)
	}
	fmt.Printf("%-18s: %s (%02X)\n", "NEW LICENSEE CODE", licensee_code_map[headers.new_licensee_code], headers.new_licensee_code)
	fmt.Printf("%-18s: %02X\n", "SGB Flag", headers.sgb_flag)
	fmt.Printf("%-18s: %s (%02X)\n", "CARTBRIDGE TYPE", cartbridge_type_map[headers.cartridge_type], headers.cartridge_type)
	fmt.Printf("%-18s: %d KiB (n. banking %d) (%d)\n", "ROM SIZE", rom_size_map[headers.rom_size].size, rom_size_map[headers.rom_size].banks, headers.rom_size)
	fmt.Printf("%-18s: %d KiB (%d)\n", "RAM SIZE", ram_size_map[headers.ram_size], headers.ram_size)
	fmt.Printf("%-18s: %s (%d)\n", "DESTINATION CODE", destination_code_map[headers.destination_code], headers.destination_code)
	fmt.Printf("---------------------------------------------------------\n")
}

func GetTitle() string {
	return headers.title
}

func IsMBC1() bool {
	return headers.cartridge_type == 0x1 || headers.cartridge_type == 0x2 || headers.cartridge_type == 0x3
}
func IsMBC3() bool {
	return headers.cartridge_type == 0x0F || headers.cartridge_type == 0x10 || headers.cartridge_type == 0x11 || headers.cartridge_type == 0x12 || headers.cartridge_type == 0x13
}
func HasBattery() bool {
	return utility.Contains(cartbridge_with_battery, uint(headers.cartridge_type))
}

func GetRomBankNumber() uint {
	return rom_size_map[headers.rom_size].banks
}
func GetRomBankBits() uint {
	return rom_size_map[headers.rom_size].bits
}
func GetRamBankNumber() uint {
	return ram_size_map[headers.ram_size] / 8
}

func IsCGB() bool {
	return headers.cgb_flag == 0xC0 || headers.cgb_flag == 0x80
}
func IsGB() bool {
	return headers.cgb_flag == 0x00 //|| headers.cgb_flag != 0xC0
}
