package ast

import (
	"testing"

	"github.com/dikaeinstein/monkey/token"
)

func TestString(t *testing.T) {
	program := Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
			},
		},
	}

	expect := "let myVar = anotherVar;"
	if program.String() != expect {
		t.Errorf("program.String() wrong. want: %q got: %q", expect, program.String())
	}
}
