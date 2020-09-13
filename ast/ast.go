package ast

import (
	"github.com/dikaeinstein/monkey/token"
)

// Node represents the AST node
type Node interface {
	TokenLiteral() string
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

// LetStatement represents the let statement node
type LetStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

// TokenLiteral returns the 'LetStatement' node token literal
func (ls *LetStatement) TokenLiteral() string {
	return ls.Token.Literal
}

func (ls *LetStatement) statementNode() {}

// Identifier represents an identifier node
type Identifier struct {
	Token token.Token
	Value string
}

// TokenLiteral returns the 'Identifier' node token literal
func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) expressionNode() {}

// ReturnStatement represents return statement node
type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression
}

// TokenLiteral returns the 'ReturnStatement' node token literal
func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReturnStatement) statementNode() {}
