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
		Make(OpConstant, 1),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
	}

	expected := `0000 OpAdd
0001 OpConstant 1
0004 OpConstant 2
0007 OpConstant 65535
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
