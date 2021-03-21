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
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "add"},
					Value: "add",
				},
				Value: &FunctionLiteral{
					Token: token.Token{Type: token.FUNCTION, Literal: "fn"},
					Parameters: []*Identifier{
						{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
						{Token: token.Token{Type: token.IDENT, Literal: "y"}, Value: "y"},
					},
					Body: &BlockStatement{
						Token: token.Token{Type: token.LBRACE, Literal: "{"},
						Statements: []Statement{
							&ReturnStatement{
								Token: token.Token{Type: token.RETURN, Literal: "return"},
								ReturnValue: &InfixExpression{
									Token: token.Token{Type: token.PLUS, Literal: "+"},
									Left: &Identifier{
										Token: token.Token{Type: token.IDENT, Literal: "x"},
										Value: "x",
									},
									Operator: string(token.PLUS),
									Right: &Identifier{
										Token: token.Token{Type: token.IDENT, Literal: "y"},
										Value: "y",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	expect := `let myVar = anotherVar;let add = fn(x, y) return (x + y);;`
	if program.String() != expect {
		t.Errorf("program.String() wrong. want: %q got: %q", expect, program.String())
	}
}
