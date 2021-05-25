package vm

import (
	"fmt"
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
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	runVMTests(t, testCases)
}

func TestBooleanExpression(t *testing.T) {
	testCases := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!(if (false) { 5; })", true},
	}

	runVMTests(t, testCases)
}

func TestConditional(t *testing.T) {
	testCases := []vmTestCase{
		{"if (true) { 10 }", 10},
		{"if (true) { 10 } else { 20 }", 10},
		{"if (false) { 10 } else { 20 } ", 20},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 }", object.NullValue()},
		{"if (false) { 10 }", object.NullValue()},
		{"if ((if (false) { 10 })) { 10 } else { 20 }", 20},
	}

	runVMTests(t, testCases)
}

func TestGlobalLetStatements(t *testing.T) {
	testCases := []vmTestCase{
		{"let one = 1; one", 1},
		{"let one = 1; let two = 2; one + two", 3},
		{"let one = 1; let two = one + one; one + two", 3},
	}

	runVMTests(t, testCases)
}

func TestStringExpressions(t *testing.T) {
	testCases := []vmTestCase{
		{`"monkey"`, "monkey"},
		{`"mon" + "key"`, "monkey"},
		{`"mon" + "key" + "banana"`, "monkeybanana"},
	}

	runVMTests(t, testCases)
}

func TestArrayLiterals(t *testing.T) {
	testCases := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []int{3, 12, 11}},
	}

	runVMTests(t, testCases)
}

func TestHashLiterals(t *testing.T) {
	testCases := []vmTestCase{
		{
			"{}", map[string]int64{},
		},
		{
			"{1: 2, 2: 3}",
			map[int]int64{
				1: 2,
				2: 3,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[int]int64{
				2: 4,
				6: 16,
			},
		},
	}

	runVMTests(t, testCases)
}

func TestIndexExpressions(t *testing.T) {
	testCases := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"[][0]", object.NullValue()},
		{"[1, 2, 3][99]", object.NullValue()},
		{"[1][-1]", object.NullValue()},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
		{"{1: 1}[0]", object.NullValue()},
		{"{}[0]", object.NullValue()},
	}

	runVMTests(t, testCases)
}

func TestCallingFunctionsWithoutArguments(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
			let fivePlusTen = fn() { 5 + 10; };
			fivePlusTen();
			`,
			expected: 15,
		},
		{
			input: `
			let one = fn() { 1; };
			let two = fn() { 2; };
			one() + two()
			`,
			expected: 3,
		},
		{
			input: `
			let a = fn() { 1 };
			let b = fn() { a() + 1 };
			let c = fn() { b() + 1 };
			c();
			`,
			expected: 3,
		},
		{
			input: `
			let earlyExit = fn() { return 99; 100; };
			earlyExit();
			`,
			expected: 99,
		},
		{
			input: `
			let earlyExit = fn() { return 99; return 100; };
			earlyExit();
			`,
			expected: 99,
		},
		{
			input: `
			let noReturn = fn() { };
			noReturn();
			`,
			expected: object.NullValue(),
		},
		{
			input: `
			let noReturn = fn() { };
			let noReturnTwo = fn() { noReturn(); };
			noReturn();
			noReturnTwo();
			`,
			expected: object.NullValue(),
		},
		{
			input: `
			let returnsOne = fn() { 1; };
			let returnsOneReturner = fn() { returnsOne; };
			returnsOneReturner()();
			`,
			expected: 1,
		},
	}

	runVMTests(t, testCases)
}

func TestCallingFunctionsWithBindings(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
			let one = fn() { let one = 1; one };
			one();
			`,
			expected: 1,
		},
		{
			input: `
			let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
			oneAndTwo();
			`,
			expected: 3,
		},
		{
			input: `
			let oneAndTwo = fn() { let one = 1; let two = 2; one + two; };
			let threeAndFour = fn() { let three = 3; let four = 4; three + four; };
			oneAndTwo() + threeAndFour();
			`,
			expected: 10,
		},
		{
			input: `
			let firstFoobar = fn() { let foobar = 50; foobar; };
			let secondFoobar = fn() { let foobar = 100; foobar; };
			firstFoobar() + secondFoobar();
			`,
			expected: 150,
		},
		{
			input: `
			let globalSeed = 50;
			let minusOne = fn() {
				let num = 1;
				globalSeed - num;
			}
			let minusTwo = fn() {
			let num = 2;
			globalSeed - num;
			}
			minusOne() + minusTwo();
			`,
			expected: 97,
		},
	}
	runVMTests(t, testCases)
}

func TestCallingFunctionsWithArgumentsAndBindings(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
			let identity = fn(a) { a; };
			identity(4);
			`,
			expected: 4,
		},
		{
			input: `
			let sum = fn(a, b) { a + b; };
			sum(1, 2);
			`,
			expected: 3,
		},
		{
			input: `
			let sum = fn(a, b) {
				let c = a + b;
				c;
			};
			sum(1, 2);
			`,
			expected: 3,
		},
		{
			input: `
			let sum = fn(a, b) {
				let c = a + b;
				c;
			};
			sum(1, 2) + sum(3, 4);
			`,
			expected: 10,
		},
		{
			input: `
			let sum = fn(a, b) {
				let c = a + b;
				c;
			};
			let outer = fn() {
			sum(1, 2) + sum(3, 4);
			};
			outer();
			`,
			expected: 10,
		},
		{
			input: `
			let globalNum = 10;
			let sum = fn(a, b) {
				let c = a + b;
				c + globalNum;
			};
			let outer = fn() {
				sum(1, 2) + sum(3, 4) + globalNum;
			};
			outer() + globalNum;
			`,
			expected: 50,
		},
	}

	runVMTests(t, testCases)
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	testCases := []vmTestCase{
		{
			input:    `fn() { 1; }(1);`,
			expected: `wrong number of arguments: want=0, got=1`,
		},
		{
			input:    `fn(a) { a; }();`,
			expected: `wrong number of arguments: want=1, got=0`,
		},
		{
			input:    `fn(a, b) { a + b; }(1);`,
			expected: `wrong number of arguments: want=2, got=1`,
		},
	}

	for _, tt := range testCases {
		program := test.Parse(tt.input)
		compiler := compile.New(compile.NewSymbolTable(), []object.Object{})

		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		vm := New(compiler.Bytecode())

		err = vm.Run()
		if err == nil {
			t.Fatalf("expected VM error but resulted in none.")
		}
		if err.Error() != tt.expected {
			t.Fatalf("wrong VM error: want=%q, got=%q", tt.expected, err)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	testCases := []vmTestCase{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{
			`len(1)`,
			object.Error("argument to `len` not supported, got INTEGER"),
		},
		{
			`len("one", "two")`,
			object.Error("wrong number of arguments. got=2, want=1"),
		},
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`puts("hello", "world!")`, object.NullValue()},
		{`first([1, 2, 3])`, 1},
		{`first([])`, object.NullValue()},
		{
			`first(1)`,
			object.Error("argument to `first` must be ARRAY, got INTEGER"),
		},
		{`last([1, 2, 3])`, 3},
		{`last([])`, object.NullValue()},
		{
			`last(1)`,
			object.Error("argument to `last` must be ARRAY, got INTEGER"),
		},
		{`rest([1, 2, 3])`, []int{2, 3}},
		{`rest([])`, object.NullValue()},
		{`push([], 1)`, []int{1}},
		{
			`push(1, 1)`,
			object.Error("argument to `push` must be ARRAY, got INTEGER"),
		},
	}

	runVMTests(t, testCases)
}

func TestClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
			let newClosure = fn(a) {
				fn() { a; };
			};
			let closure = newClosure(99);
			closure();
			`,
			expected: 99,
		},
		{
			input: `
			let newAdder = fn(a, b) {
				fn(c) { a + b + c };
			};
			let adder = newAdder(1, 2);
			adder(8);
			`,
			expected: 11,
		},
		{
			input: `
			let newAdder = fn(a, b) {
				let c = a + b;
				fn(d) { c + d };
			};
			let adder = newAdder(1, 2);
			adder(8);
			`,
			expected: 11,
		},
	}

	runVMTests(t, tests)
}

func TestRecursiveFunctions(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
			let countDown = fn(x) {
				if (x == 0) {
					return 0;
				} else {
					countDown(x - 1);
				}
			};
			countDown(1);
			`,
			expected: 0,
		},
		{
			input: `
			let countDown = fn(x) {
				if (x == 0) {
					return 0;
				} else {
					countDown(x - 1);
				}
			};
			let wrapper = fn() {
				countDown(1);
			};
			wrapper();
			`,
			expected: 0,
		},
		{
			input: `
			let wrapper = fn() {
				let countDown = fn(x) {
					if (x == 0) {
						return 0;
					} else {
						countDown(x - 1);
					}
				};
				countDown(1);
			};
			wrapper();
			`,
			expected: 0,
		},
	}

	runVMTests(t, testCases)
}

func TestRecursiveFibonacci(t *testing.T) {
	testCases := []vmTestCase{
		{
			input: `
			let fibonacci = fn(x) {
				if (x == 0) {
					return 0;
				} else {
					if (x == 1) {
						return 1;
					} else {
						fibonacci(x - 1) + fibonacci(x - 2);
					}
				}
			};

			fibonacci(15);
			`,
			expected: 610,
		},
	}

	runVMTests(t, testCases)
}

func runVMTests(t *testing.T, testCases []vmTestCase) {
	t.Helper()

	for _, tC := range testCases {
		program := test.Parse(tC.input)

		compiler := compile.NewCompilerWithBuiltins([]object.Object{})
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		for i, constant := range compiler.Bytecode().Constants {
			fmt.Printf("CONSTANT %d %p (%T):\n", i, constant, constant)
			switch constant := constant.(type) {
			case *object.CompiledFunction:
				fmt.Printf(" Instructions:\n%s", constant.Instructions)
			case object.Integer:
				fmt.Printf(" Value: %d\n", constant)
			}
			fmt.Printf("\n")
		}

		vm := New(compiler.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, tC.expected, stackElem)
	}
}

//gocyclo:ignore
func testExpectedObject(t *testing.T, expected interface{}, actual object.Object) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		err := test.IntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}
	case bool:
		err := testBooleanObject(expected, actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}
	case *object.Null:
		if actual != object.NullValue() {
			t.Errorf("object is not Null: %T (%+v)", actual, actual)
		}
	case []int:
		array, ok := actual.(*object.Array)
		if !ok {
			t.Errorf("object not Array: %T (%+v)", actual, actual)
			return
		}
		if len(array.Elements) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d",
				len(expected), len(array.Elements))
			return
		}
		for i, expectedElem := range expected {
			err := test.IntegerObject(int64(expectedElem), array.Elements[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	case map[string]int64:
		hash, ok := actual.(*object.Hash)
		if !ok {
			t.Errorf("object is not Hash. got=%T (%+v)", actual, actual)
			return
		}

		if len(hash.Pairs) != len(expected) {
			t.Errorf("hash has wrong number of Pairs. want=%d, got=%d",
				len(expected), len(hash.Pairs))
			return
		}

		for expectedKey, expectedValue := range expected {
			value, ok := hash.Pairs[object.String(expectedKey)]
			if !ok {
				t.Errorf("no pair for given key in Pairs")
			}
			err := test.IntegerObject(expectedValue, value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}
	}
}

func testBooleanObject(expected bool, actual object.Object) error {
	result, ok := actual.(object.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)",
			actual, actual)
	}

	if bool(result) != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t",
			result, expected)
	}
	return nil
}
