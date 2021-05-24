package eval

import (
	"testing"

	"github.com/dikaeinstein/monkey/lexer"
	"github.com/dikaeinstein/monkey/object"
	"github.com/dikaeinstein/monkey/parser"
)

func TestEvalIntegerExpression(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tC := range testCases {
		evaluated := testEval(t, tC.input)
		testIntegerObject(t, evaluated, tC.expected)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!"`
	evaluated := testEval(t, input)
	str, ok := evaluated.(object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}
	if str != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`
	evaluated := testEval(t, input)
	str, ok := evaluated.(object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}
	if str != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
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
	}

	for _, tC := range testCases {
		evaluated := testEval(t, tC.input)
		testBooleanObject(t, evaluated, tC.expected)
	}
}

func TestBangOperator(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}
	for _, tC := range testCases {
		evaluated := testEval(t, tC.input)
		testBooleanObject(t, evaluated, tC.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	testCases := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
	}

	for _, tC := range testCases {
		evaluated := testEval(t, tC.input)
		integer, ok := tC.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
	}

	for _, tC := range testCases {
		evaluated := testEval(t, tC.input)
		testIntegerObject(t, evaluated, tC.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	testCases := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
	if (10 > 1) {
		if (10 > 1) {
			return true + false;
		}
		return 1;
	}
	`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
		{
			`{"name": "Monkey"}[fn(x) { x }];`,
			"unusable as hash key: FUNCTION",
		},
	}

	for _, tC := range testCases {
		evaluated := testEval(t, tC.input)
		errVal, ok := evaluated.(object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
		}
		if string(errVal) != tC.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tC.expectedMessage, errVal)
		}
	}
}

func TestLetStatements(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
	}

	for _, tC := range testCases {
		testIntegerObject(t, testEval(t, tC.input), tC.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"
	evaluated := testEval(t, input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}
	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}
	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}
	expectedBody := "(x + 2)"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	testCases := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}

	for _, tC := range testCases {
		testIntegerObject(t, testEval(t, tC.input), tC.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
	let fibonacci = fn(x) {
		if (x == 0) {
			0
		} else {
			if (x == 1) {
				return 1;
			} else {
				return fibonacci(x - 1) + fibonacci(x - 2);
			}
		}
	};

	fibonacci(15);
	`

	testIntegerObject(t, testEval(t, input), 610)
}

func TestBuiltinFunctions(t *testing.T) {
	testCases := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
		{`len([1,2,3])`, 3},
		{`len([])`, 0},
		{`first([1, "2", "a"])`, 1},
		{`first([])`, object.NullValue()},
		{`rest([1, 2, 3, 4, 5])`, "[2, 3, 4, 5]"},
		{`rest([1])`, "[]"},
		{`rest([])`, object.NullValue()},
		{`let a = [1, 2, 3, 4]; rest(a)`, "[2, 3, 4]"},
		{`push([1, 2, 3], 4)`, "[1, 2, 3, 4]"},
		{`push([], 4)`, "[4]"},
	}

	for _, tC := range testCases {
		evaluated := testEval(t, tC.input)
		switch expected := tC.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			switch obj := evaluated.(type) {
			case object.Error:
				if string(obj) != expected {
					t.Errorf("wrong error message. expected=%q, got=%q",
						expected, obj)
				}
			case *object.Array:
				if obj.Inspect() != expected {
					t.Errorf("expected %s, got %s", expected, obj.Inspect())
				}
			default:
				t.Errorf("unknown Type. got=%T (%+v))", evaluated, evaluated)
			}
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(t, input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d",
			len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	testCases := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
	}

	for _, tC := range testCases {
		evaluated := testEval(t, tC.input)
		integer, ok := tC.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestHashLiterals(t *testing.T) {
	input := `let two = "two";
	{
		"one": 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 / 2,
		4: 4,
		true: 5,
		false: 6
	}`

	evaluated := testEval(t, input)
	result, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[string]int64{
		"one":   1,
		"two":   2,
		"three": 3,
		"4":     4,
		"true":  5,
		"false": 6,
	}
	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs. got=%d", len(result.Pairs))
	}
	for expectedKey, expectedValue := range expected {
		value, ok := result.Pairs[object.String(expectedKey)]
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}
		testIntegerObject(t, value, expectedValue)
	}
}

func TestHashIndexExpressions(t *testing.T) {
	testCases := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5}[false]`,
			5,
		},
	}

	for _, tC := range testCases {
		evaluated := testEval(t, tC.input)
		integer, ok := tC.expected.(int)

		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestQuote(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			`quote(5)`,
			`5`,
		},
		{
			`quote(5 + 8)`,
			`(5 + 8)`,
		},
		{
			`quote(foobar)`,
			`foobar`,
		},
		{
			`quote(foobar + barfoo)`,
			`(foobar + barfoo)`,
		},
	}

	for _, tC := range testCases {
		evaluated := testEval(t, tC.input)
		quote, ok := evaluated.(*object.Quote)
		if !ok {
			t.Fatalf("expected *object.Quote. got=%T (%+v)",
				evaluated, evaluated)
		}
		if quote.Node == nil {
			t.Fatalf("quote.Node is nil")
		}
		if quote.Node.String() != tC.expected {
			t.Errorf("not equal. got=%q, want=%q",
				quote.Node.String(), tC.expected)
		}
	}
}

func TestQuoteUnquote(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{
			`quote(unquote(4))`,
			`4`,
		},
		{
			`quote(unquote(4 + 4))`,
			`8`,
		},
		{
			`quote(8 + unquote(4 + 4))`,
			`(8 + 8)`,
		},
		{
			`quote(unquote(4 + 4) + 8)`,
			`(8 + 8)`,
		},
		{
			`let foobar = 8;
			quote(foobar)`,
			`foobar`,
		},
		{
			`let foobar = 8;
			quote(unquote(foobar))`,
			`8`,
		},
		{
			`quote(unquote(true))`,
			`true`,
		},
		{
			`quote(unquote(true == false))`,
			`false`,
		},
		{
			`quote(unquote(quote(4 + 4)))`,
			`(4 + 4)`,
		},
		{
			`let quotedInfixExpression = quote(4 + 4);
			quote(unquote(4 + 4) + unquote(quotedInfixExpression))`,
			`(8 + (4 + 4))`,
		},
	}

	for _, tC := range testCases {
		evaluated := testEval(t, tC.input)
		quote, ok := evaluated.(*object.Quote)
		if !ok {
			t.Fatalf("expected *object.Quote. got=%T (%+v)",
				evaluated, evaluated)
		}
		if quote.Node == nil {
			t.Fatalf("quote.Node is nil")
		}
		if quote.Node.String() != tC.expected {
			t.Errorf("not equal. got=%q, want=%q",
				quote.Node.String(), tC.expected)
		}
	}
}

func testNullObject(t *testing.T, obj object.Object) {
	if obj != object.NullValue() {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
	}
}

func testEval(t *testing.T, input string) object.Object {
	t.Helper()

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	env := object.NewEnvironment()

	return Eval(program, env)
}

func checkParserErrors(t *testing.T, p *parser.Parser) {
	t.Helper()

	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has had %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) {
	t.Helper()

	intObj, ok := obj.(object.Integer)
	if !ok {
		t.Errorf("object is not an Integer. got=%T (%+v)", obj, obj)
	}
	result := int64(intObj)
	if result != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result, expected)
	}
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) {
	t.Helper()

	boolObj, ok := obj.(object.Boolean)
	if !ok {
		t.Errorf("object is not a Boolean. got=%T (%+v)", obj, obj)
	}
	result := bool(boolObj)
	if result != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result, expected)
	}
}
