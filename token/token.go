package token

// Type is the set of lexical token types of the Monkey programming language.
type Type string

// Token represents a valid token of the monkey language
type Token struct {
	Type
	Literal string
}

const (
	ILLEGAL Type = "ILLEGAL"
	EOF     Type = "EOF"

	// Identifiers + literals
	IDENT  Type = "IDENT" // add, foobar, x, y, ...
	INT    Type = "INT"
	STRING Type = "STRING"

	// Operators
	ASSIGN   Type = "="
	PLUS     Type = "+"
	MINUS    Type = "-"
	BANG     Type = "!"
	ASTERISK Type = "*"
	SLASH    Type = "/"
	LT       Type = "<"
	GT       Type = ">"
	EQ       Type = "=="
	NotEQ    Type = "!="

	// Delimiters
	COMMA     Type = ","
	SEMICOLON Type = ";"
	LPAREN    Type = "("
	RPAREN    Type = ")"
	LBRACE    Type = "{"
	RBRACE    Type = "}"
	LBRACKET  Type = "["
	RBRACKET  Type = "]"
	COLON     Type = ":"

	// Keywords
	FUNCTION Type = "FUNCTION"
	LET      Type = "LET"
	IF       Type = "IF"
	ELSE     Type = "ELSE"
	RETURN   Type = "RETURN"
	FALSE    Type = "FALSE"
	TRUE     Type = "TRUE"
)

var keywords = map[string]Type{
	"fn":     FUNCTION,
	"let":    LET,
	"if":     IF,
	"else":   ELSE,
	"return": RETURN,
	"true":   TRUE,
	"false":  FALSE,
}

// LookupIdent returns the appropriate keyword token type or IDENT
func LookupIdent(ident string) Type {
	if tokenType, ok := keywords[ident]; ok {
		return tokenType
	}
	return IDENT
}
