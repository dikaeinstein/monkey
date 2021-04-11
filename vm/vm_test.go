package vm

import (
	"testing"

	"github.com/dikaeinstein/monkey/compile"
	"github.com/dikaeinstein/monkey/object"
	"github.com/dikaeinstein/monkey/test"
)

type vmTestCase struct {
	input    string
	expected interface{}
}

func TestIntegerArithmetic(t *testing.T) {
	testCases := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
	}

	runVMTests(t, testCases)
}

func runVMTests(t *testing.T, testCases []vmTestCase) {
	t.Helper()

	for _, tC := range testCases {
		program := test.Parse(tC.input)

		compiler := compile.New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(compiler.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.StackTop()
		testExpectedObject(t, tC.expected, stackElem)
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := test.TestIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	}
}
