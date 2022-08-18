package ppu

var ReadFromMemory func(addr uint, bytes ...uint) uint

var dma_old uint
var dma_delay uint
var dma_tranferring bool
var current_byte uint

func DMAStart() {
	if !dma_tranferring {
		dma_old = GetDMA() << 8
		// max is 0xDF00 so in case just decrease
		if dma_old > 0xDF00 {
			dma_old -= 0x2000
		}
		current_byte = 0
		dma_tranferring = true
		dma_delay = 1
	}
}

func DMATransfer() {
	b := byte(ReadFromMemory(dma_old + current_byte))

	WriteToOAMMemory(0xFE00+current_byte, b)

	current_byte++

	// since dma transfers to 0xFE00 - 0xFE9F ( so 0x9F bytes )
	dma_tranferring = current_byte <= 0x9F
}

func DMATick() {
	if !dma_tranferring {
		return
	}
	if dma_delay > 0 {
		dma_delay--
		return
	}

	DMATransfer()
}

func GetDMATransferring() bool {
	return dma_tranferring
}
