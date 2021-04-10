package ast

import (
	"reflect"
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

func TestModify(t *testing.T) {
	one := func() Expression { return &IntegerLiteral{Value: 1} }
	two := func() Expression { return &IntegerLiteral{Value: 2} }

	turnOneIntoTwo := func(node Node) Node {
		integer, ok := node.(*IntegerLiteral)
		if !ok {
			return node
		}
		if integer.Value != 1 {
			return node
		}
		integer.Value = 2
		return integer
	}

	testCases := []struct {
		input    Node
		expected Node
	}{
		{
			one(),
			two(),
		},
		{
			&Program{
				Statements: []Statement{
					&ExpressionStatement{Expression: one()},
				},
			},
			&Program{
				Statements: []Statement{
					&ExpressionStatement{Expression: two()},
				},
			},
		},
		{
			&InfixExpression{Left: one(), Operator: "+", Right: two()},
			&InfixExpression{Left: two(), Operator: "+", Right: two()},
		},
		{
			&InfixExpression{Left: two(), Operator: "+", Right: one()},
			&InfixExpression{Left: two(), Operator: "+", Right: two()},
		},
		{
			&PrefixExpression{Operator: "-", Right: one()},
			&PrefixExpression{Operator: "-", Right: two()},
		},
		{
			&IndexExpression{Left: one(), Index: one()},
			&IndexExpression{Left: two(), Index: two()},
		},
		{
			&IfExpression{
				Condition: one(),
				Consequence: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{Expression: one()},
					},
				},
				Alternative: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{Expression: one()},
					},
				},
			},
			&IfExpression{
				Condition: two(),
				Consequence: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{Expression: two()},
					},
				},
				Alternative: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{Expression: two()},
					},
				},
			},
		},
		{
			&ReturnStatement{ReturnValue: one()},
			&ReturnStatement{ReturnValue: two()},
		},
		{
			&LetStatement{Value: one()},
			&LetStatement{Value: two()},
		},
		{
			&FunctionLiteral{
				Parameters: []*Identifier{},
				Body: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{Expression: one()},
					},
				},
			},
			&FunctionLiteral{
				Parameters: []*Identifier{},
				Body: &BlockStatement{
					Statements: []Statement{
						&ExpressionStatement{Expression: two()},
					},
				},
			},
		},
		{
			&ArrayLiteral{Elements: []Expression{one(), one()}},
			&ArrayLiteral{Elements: []Expression{two(), two()}},
		},
	}

	for _, tC := range testCases {
		modified := Modify(tC.input, turnOneIntoTwo)
		equal := reflect.DeepEqual(modified, tC.expected)
		if !equal {
			t.Errorf("not equal. got=%#v, want=%#v",
				modified, tC.expected)
		}
	}

	hashLiteral := &HashLiteral{
		Pairs: map[Expression]Expression{
			one(): one(),
			one(): one(),
		},
	}
	Modify(hashLiteral, turnOneIntoTwo)
	for key, val := range hashLiteral.Pairs {
		key, _ := key.(*IntegerLiteral)
		if key.Value != 2 {
			t.Errorf("value is not %d, got=%d", 2, key.Value)
		}
		val, _ := val.(*IntegerLiteral)
		if val.Value != 2 {
			t.Errorf("value is not %d, got=%d", 2, val.Value)
		}
	}
}
