package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type Opcode byte

const (
	OpConstant Opcode = iota
	OpPop
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpGreaterThan
	OpBang
	OpMinus
	OpJump
	OpJumpNotTruthy
	OpNull
	OpGetGlobal
	OpSetGlobal
	OpArray
	OpHash
	OpIndex
	OpCall
	OpReturn
	OpReturnValue
	OpGetLocal
	OpSetLocal
	OpGetBuiltin
	OpClosure
	OpGetFree
	OpCurrentClosure
)

// OperandWidth is the number of bytes an operand takes up
const (
	_ = iota
	OperandWidth1
	OperandWidth2
)

const (
	_ = iota
	operandCount1
	operandCount2
)

type Instructions []byte

func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(Opcode(ins[i]))
		if err != nil {
			fmt.Fprintf(&out, "Error: %s\n", err)
			continue
		}

		operands, n := ReadOperands(def, ins[i+1:])
		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + int(n)
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case operandCount1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case operandCount2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	default:
		return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
	}
}

// Definition helps make an Opcode readable
type Definition struct {
	// Name of Opcode
	Name string
	// OperandWidths contains the OperandWidth for each operand
	OperandWidths []uint
}

var definitions = map[Opcode]*Definition{
	OpConstant:       {Name: "OpConstant", OperandWidths: []uint{OperandWidth2}},
	OpPop:            {Name: "OpPop"},
	OpAdd:            {Name: "OpAdd"},
	OpSub:            {Name: "OpSub"},
	OpMul:            {Name: "OpMul"},
	OpDiv:            {Name: "OpDiv"},
	OpTrue:           {Name: "OpTrue"},
	OpFalse:          {Name: "OpFalse"},
	OpEqual:          {Name: "OpEqual"},
	OpNotEqual:       {Name: "OpNotEqual"},
	OpGreaterThan:    {Name: "OpGreaterThan"},
	OpBang:           {Name: "OpBang"},
	OpMinus:          {Name: "OpMinus"},
	OpJump:           {Name: "OpJump", OperandWidths: []uint{OperandWidth2}},
	OpJumpNotTruthy:  {Name: "OpJumpNotTruthy", OperandWidths: []uint{OperandWidth2}},
	OpNull:           {Name: "OpNull"},
	OpGetGlobal:      {Name: "OpGetGlobal", OperandWidths: []uint{OperandWidth2}},
	OpSetGlobal:      {Name: "OpSetGlobal", OperandWidths: []uint{OperandWidth2}},
	OpArray:          {Name: "OpArray", OperandWidths: []uint{OperandWidth2}},
	OpHash:           {Name: "OpHash", OperandWidths: []uint{OperandWidth2}},
	OpIndex:          {Name: "OpIndex"},
	OpCall:           {Name: "OpCall", OperandWidths: []uint{OperandWidth1}},
	OpReturnValue:    {Name: "OpReturnValue"},
	OpReturn:         {Name: "OpReturn"},
	OpGetLocal:       {Name: "OpGetLocal", OperandWidths: []uint{OperandWidth1}},
	OpSetLocal:       {Name: "OpSetLocal", OperandWidths: []uint{OperandWidth1}},
	OpGetBuiltin:     {Name: "OpGetBuiltin", OperandWidths: []uint{OperandWidth1}},
	OpClosure:        {Name: "OpClosure", OperandWidths: []uint{OperandWidth2, OperandWidth1}},
	OpGetFree:        {"OpGetFree", []uint{OperandWidth1}},
	OpCurrentClosure: {Name: "OpCurrentClosure"},
}

func Lookup(op Opcode) (*Definition, error) {
	def, ok := definitions[op]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}

	return def, nil
}

// Make creates a single bytecode instruction
// which includes the Opcode and it's operands
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	var instructionLen uint = 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	var offset uint = 1
	for i, o := range operands {
		width := def.OperandWidths[i]

		switch width {
		case OperandWidth2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case OperandWidth1:
			instruction[offset] = byte(o)
		}

		offset += width
	}

	return instruction
}

// ReadOperands reverses a bytecode instruction and reads its operands.
// Returning the operands in the instruction and n number of bytes read.
func ReadOperands(def *Definition, ins []byte) (operandsRead []int, n uint) {
	operandsRead = make([]int, len(def.OperandWidths))

	var offset uint
	for i, width := range def.OperandWidths {
		switch width {
		case OperandWidth2:
			operandsRead[i] = int(ReadUint16(ins[offset:]))
		case OperandWidth1:
			operandsRead[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}

	n = offset
	return operandsRead, n
}

func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

func ReadUint8(ins Instructions) uint8 { return ins[0] }
