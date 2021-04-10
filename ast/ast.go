package ast

import (
	"bytes"
	"strings"

	"github.com/dikaeinstein/monkey/token"
)

// Node represents the AST node
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement describes a statement node
type Statement interface {
	Node
	statementNode()
}

// Expression describes an expression node
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of the AST
type Program struct {
	Statements []Statement
}

// TokenLiteral returns the root node token literal
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}

	return ""
}
func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

// LetStatement represents the let statement node
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier
	Value Expression
}

// TokenLiteral returns the 'LetStatement' node token literal
func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer

	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.Value)
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

// Identifier represents an identifier node
type Identifier struct {
	Token token.Token
	Value string
}

// TokenLiteral returns the 'Identifier' node token literal
func (i *Identifier) expressionNode()      {}
func (i *Identifier) String() string       { return i.Value }
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// ReturnStatement represents return statement node
type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression
}

// TokenLiteral returns the 'ReturnStatement' node token literal
func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

// ExpressionStatement represents valid expressions statement node. ex: 2 + 2
type ExpressionStatement struct {
	Expression
	Token token.Token // the first token in the expression
}

// TokenLiteral returns the 'ExpressionStatement' node token literal
func (ex *ExpressionStatement) TokenLiteral() string {
	return ex.Token.Literal
}
func (ex *ExpressionStatement) statementNode() {}
func (ex *ExpressionStatement) String() string {
	if ex.Expression != nil {
		return ex.Expression.String()
	}

	return ""
}

// IntegerLiteral represents a integer node
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// PrefixExpression represents a prefix expression e.g: -5 or !false
type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

// InfixExpression represents an infix expression e.g: 1 + 2
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Right    Expression
	Operator string
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

// Boolean represents a boolean literal value
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bl *BlockStatement) statementNode()       {}
func (bl *BlockStatement) TokenLiteral() string { return bl.Token.Literal }
func (bl *BlockStatement) String() string {
	var out bytes.Buffer

	for _, s := range bl.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (i *IfExpression) expressionNode()      {}
func (i *IfExpression) TokenLiteral() string { return i.Token.Literal }
func (i *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(i.Condition.String())
	out.WriteString(" ")
	out.WriteString(i.Consequence.String())

	if i.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(i.Alternative.String())
	}

	return out.String()
}

// FunctionLiteral represents a function expression
type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString(" ")
	out.WriteString(fl.Body.String())

	return out.String()
}

type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

type ArrayLiteral struct {
	Token    token.Token // The [ token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type IndexExpression struct {
	Token token.Token // The [ token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}

type HashLiteral struct {
	Token token.Token // the '{' token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// MacroLiteral represents a function expression
type MacroLiteral struct {
	Token      token.Token // The 'macro' token
	Parameters []*Identifier
	Body       *BlockStatement
}

func (ml *MacroLiteral) expressionNode()      {}
func (ml *MacroLiteral) TokenLiteral() string { return ml.Token.Literal }
func (ml *MacroLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range ml.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("macro")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	out.WriteString(" ")
	out.WriteString(ml.Body.String())

	return out.String()
}
