package ppu

import "fmt"

func Test() {

	SetSTATModeHBlank()
	fmt.Println(GetModeSTAT())
	SetSTATModeVBlank()
	fmt.Println(GetModeSTAT())
	SetSTATModeOAM()
	fmt.Println(GetModeSTAT())
}
