package decoder

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	// "strconv"
	"strings"

	mmu "github.com/giammirove/gbemu/libs/mmu"
)

const OPCODE_JSON_PATH = "./assets/opcodes.json"
const OPCODE_PREFIX = 0xCB

// d8 in opcodes.json is the input -> d8 == n,#
// -> one byte signed immediate value
// a16,d16 -> nn -> two byte immediate value.

type Operand_t struct {
	Name      string
	Immediate bool
	Decrement bool
	Increment bool
	Bytes     uint
	Withvalue bool
	Value     uint
}

type Instruction_t struct {
	Opcode    uint
	Immediate bool
	Operands  []Operand_t
	Cycles    []uint
	Bytes     uint
	Flags     map[string]string
	Mnemonic  string
	Prefixed  bool
}

type Instructions_t struct {
	Unprefixed map[string]Instruction_t `json:"unprefixed"`
	Cbprefixed map[string]Instruction_t `json:"cbprefixed"`
}

var instructions Instructions_t
var data []byte
var Cycle func(val uint)

func InitDecoder() {
	// Read_Headers("./roms/tetris.gb")
	ReadOpcodes()
}

func ReadOpcodes() {
	byteValue, err_read :=
		ioutil.ReadFile(OPCODE_JSON_PATH)
	if err_read != nil {
		log.Fatal(err_read)
	}

	err := json.Unmarshal(byteValue, &instructions)
	if err != nil {
		log.Fatal(err)
	}

}

func read_from_data(addr uint, args ...uint) uint {

	if len(args) > 1 {
		log.Fatal("Too many parameters")
	}

	length := uint(1)
	if len(args) == 1 {
		length = args[0]
	}

	end := int(addr + length)
	if end < 0 || end >= len(data) {
		log.Fatal("Out of bound")
	}

	var arr []byte
	for i := 0; i < int(length); i++ {
		arr = append(arr, data[int(addr)+i])
	}
	for i := length; i < 4; i++ {
		arr = append(arr, 0x00)
	}
	return uint(binary.LittleEndian.Uint32(arr))
}

func copyInstruction(dst *Instruction_t, src Instruction_t) {
	dst.Bytes = src.Bytes
	dst.Mnemonic = src.Mnemonic
	dst.Prefixed = src.Prefixed
	dst.Cycles = src.Cycles
	dst.Immediate = src.Immediate
	// copier.Copy(&dst.Operands, src.Operands)
	dst.Operands = append(dst.Operands, src.Operands...)
}

func printOperand(operand Operand_t) string {
	s := ""
	if operand.Withvalue {
		if operand.Bytes > 0 {
			s += fmt.Sprintf("0x%X", operand.Value)
		} else {
			s += fmt.Sprintf("%s", operand.Name)
		}
	} else {
		s += operand.Name
	}
	if operand.Decrement {
		s += "-"
	}
	if operand.Increment {
		s += "+"
	}
	if operand.Immediate {
		return s
	}
	return fmt.Sprintf("(%s)", s)
}

func PrintInstrunction(ins Instruction_t) string {
	s := fmt.Sprintf("(%02X) %s ", ins.Opcode, ins.Mnemonic)
	for i := 0; i < len(ins.Operands); i++ {
		s += fmt.Sprintf("%s", printOperand(ins.Operands[i]))
		if i < len(ins.Operands)-1 {
			s += ","
		}
	}
	return s
}

func Decode(addr *uint) Instruction_t {

	Cycle(4)
	opcode := mmu.ReadFromMemory(*addr)
	(*addr)++

	instruction := Instruction_t{}
	if opcode == OPCODE_PREFIX {
		Cycle(4)
		opcode = mmu.ReadFromMemory(*addr)
		opcode_conv := "0x" + strings.ToUpper(fmt.Sprintf("%02x", int(opcode)))
		(*addr)++
		copyInstruction(&instruction,
			instructions.Cbprefixed[opcode_conv])
		instruction.Prefixed = true
	} else {
		opcode_conv := "0x" + strings.ToUpper(fmt.Sprintf("%02x", int(opcode)))
		copyInstruction(&instruction, instructions.Unprefixed[opcode_conv])
	}

	if instruction.Mnemonic == "" {
		log.Fatal("Opcodes not found")
	}

	instruction.Opcode = uint(opcode)

	operands := instruction.Operands
	for i := 0; i < len(operands); i++ {
		op := operands[i]
		if op.Bytes > 0 {
			Cycle(4 * uint(op.Bytes))
			value := mmu.ReadFromMemory(*addr, op.Bytes)
			(*addr) += op.Bytes
			op.Value = value
			op.Withvalue = true
		} else {
			op.Bytes = 1
		}
		operands[i] = op
	}

	return instruction
}
