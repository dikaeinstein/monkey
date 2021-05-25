package code

import (
	"testing"
)

func TestMake(t *testing.T) {
	testCases := []struct {
		op       Opcode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
		{OpPop, []int{}, []byte{byte(OpPop)}},
		{OpSub, []int{}, []byte{byte(OpSub)}},
		{OpMul, []int{}, []byte{byte(OpMul)}},
		{OpDiv, []int{}, []byte{byte(OpDiv)}},
		{OpTrue, []int{}, []byte{byte(OpTrue)}},
		{OpFalse, []int{}, []byte{byte(OpFalse)}},
		{OpEqual, []int{}, []byte{byte(OpEqual)}},
		{OpNotEqual, []int{}, []byte{byte(OpNotEqual)}},
		{OpGreaterThan, []int{}, []byte{byte(OpGreaterThan)}},
		{OpBang, []int{}, []byte{byte(OpBang)}},
		{OpMinus, []int{}, []byte{byte(OpMinus)}},
		{OpGetLocal, []int{255}, []byte{byte(OpGetLocal), 255}},
		{OpSetLocal, []int{255}, []byte{byte(OpSetLocal), 255}},
		{OpSetLocal, []int{255}, []byte{byte(OpSetLocal), 255}},
		{OpClosure, []int{65534, 255}, []byte{byte(OpClosure), 255, 254, 255}},
	}

	for _, tC := range testCases {
		instruction := Make(tC.op, tC.operands...)

		if len(instruction) != len(tC.expected) {
			t.Errorf("instruction has wrong length. want=%d, got=%d",
				len(tC.expected), len(instruction))
		}

		for i, b := range tC.expected {
			if instruction[i] != tC.expected[i] {
				t.Errorf("wrong byte at pos %d. want=%d, got=%d",
					i, b, instruction[i])
			}
		}
	}
}

func TestInstructionsString(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpPop),
		Make(OpConstant, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
		Make(OpGetLocal, 255),
		Make(OpSetLocal, 255),
		Make(OpClosure, 65535, 255),
	}

	expected := `0000 OpAdd
0001 OpPop
0002 OpConstant 1
0005 OpConstant 2
0008 OpConstant 65535
0011 OpGetLocal 255
0013 OpSetLocal 255
0015 OpClosure 65535 255
`

	concatted := Instructions{}
	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted.\nwant=%q\ngot=%q",
			expected, concatted.String())
	}
}

func TestReadOperands(t *testing.T) {
	testCases := []struct {
		op        Opcode
		operands  []int
		bytesRead uint
	}{
		{OpConstant, []int{65535}, 2},
		{OpGetLocal, []int{255}, 1},
	}

	for _, tC := range testCases {
		instruction := Make(tC.op, tC.operands...)

		def, err := Lookup(tC.op)
		if err != nil {
			t.Fatalf("definition not found: %q\n", err)
		}

		operandsRead, n := ReadOperands(def, instruction[1:])
		if n != tC.bytesRead {
			t.Fatalf("n wrong. want=%d, got=%d", tC.bytesRead, n)
		}

		for i, want := range tC.operands {
			if operandsRead[i] != want {
				t.Errorf("operand wrong. want=%d, got=%d", want, operandsRead[i])
			}
		}
	}
}
