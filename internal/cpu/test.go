package cpu

import (
	"fmt"

	decoder "github.com/giammirove/gbemu/internal/decoder"
	registers "github.com/giammirove/gbemu/internal/registers"
)

func Test() {
	fmt.Printf("C Flag : %t\n", registers.C_flag())
	// registers.SetA(0x39)
	// registers.SetB(0x48)
	operand1 := decoder.Operand_t{Name: "HL", Immediate: false}
	operand2 := decoder.Operand_t{Name: "SP", Increment: false, Immediate: true}
	operand3 := decoder.Operand_t{Name: "a8", Withvalue: true, Value: 23, Immediate: true}
	instruction := decoder.Instruction_t{}
	instruction.Opcode = 0xF8
	instruction.Mnemonic = "LD"
	instruction.Operands = append(instruction.Operands, operand1)
	instruction.Operands = append(instruction.Operands, operand2)
	instruction.Operands = append(instruction.Operands, operand3)
	// fmt.Println(instruction)
	fmt.Println("Before")
	fmt.Printf("OP1 : %08b\n", registers.Get(operand1.Name)())
	fmt.Printf("OP2 : %08b\n", registers.Get(operand2.Name)())
	fmt.Printf("%s -> %d\n", operand2.Name, registers.Get(operand2.Name)())
	execute(instruction)
	fmt.Printf("OP1 : %08b\n", registers.Get(operand1.Name)())
	fmt.Printf("OP2 : %08b\n", registers.Get(operand2.Name)())
	fmt.Printf("H Flag : %t\n", registers.H_flag())
	fmt.Printf("C Flag : %t\n", registers.C_flag())
}
