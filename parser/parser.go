package parser

import (
	"github.com/dikaeinstein/monkey/ast"
	"github.com/dikaeinstein/monkey/lexer"
	"github.com/dikaeinstein/monkey/token"
)

// Parser is a recursive decent parser that analyses the sequence of
// tokens to generae the AST
type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
}

// New initialize and returns a new parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseProgram recursively parses the tokens and returns the root node
func (p *Parser) ParseProgram() *ast.Program {
	program := ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}

	p.nextToken()
	return nil
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	default:
		return nil
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	// TODO: We're skipping the expressions until we
	// encounter a semicolon
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) expectPeek(t token.Type) bool {
	if p.curToken.Type != t {
		return false
	}

	return true
}
