package test

import (
	"fmt"

	"github.com/dikaeinstein/monkey/ast"
	"github.com/dikaeinstein/monkey/lexer"
	"github.com/dikaeinstein/monkey/object"
	"github.com/dikaeinstein/monkey/parser"
)

func Parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func IntegerObject(expected int64, actual object.Object) error {
	result, ok := actual.(object.Integer)
	if !ok {
		return fmt.Errorf("object is not Integer. got=%T (%+v)",
			actual, actual)
	}

	if int64(result) != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d",
			result, expected)
	}

	return nil
}
