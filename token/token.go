package token

// Type is the token type
type Type string

// Token represents a valid token of the monkey language
type Token struct {
	Type    Type
	Literal string
}

const (
	ILLEGAL Type = "ILLEGAL"
	EOF          = "EOF"
	// Identifiers + literals
	// IDENT is the identifier token
	IDENT = "IDENT" // add, foobar, x, y, ...
	INT   = "INT"
	// Operators
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	LT       = "<"
	GT       = ">"
	EQ       = "=="
	NOT_EQ   = "!="
	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	// Keywords
	// 1343456
	FUNCTION = "FUNCTION"
	LET      = "LET"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	FALSE    = "FALSE"
	TRUE     = "TRUE"
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
