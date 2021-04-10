package object

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/dikaeinstein/monkey/ast"
)

type Type string

// Object represents values in the host language(GO).
// An object can either be a primitive or reference value.
type Object interface {
	Type() Type
	Inspect() string
}

const (
	ARRAY    Type = "ARRAY"
	BOOLEAN  Type = "BOOLEAN"
	BUILTIN  Type = "BUILTIN"
	ERROR    Type = "ERROR"
	FUNCTION Type = "FUNCTION"
	HASH     Type = "HASH"
	INTEGER  Type = "INTEGER"
	NULL     Type = "NULL"
	STRING   Type = "STRING"
	QUOTE    Type = "QUOTE"
	MACRO    Type = "MACRO"
)

type Integer int64

func (i Integer) Type() Type      { return INTEGER }
func (i Integer) Inspect() string { return fmt.Sprint(i) }

type String string

func (s String) Type() Type      { return STRING }
func (s String) Inspect() string { return string(s) }

type Boolean bool

func (b Boolean) Type() Type      { return BOOLEAN }
func (b Boolean) Inspect() string { return fmt.Sprint(b) }

type Null struct{}

func (n Null) Type() Type      { return NULL }
func (n Null) Inspect() string { return "null" }

type Error string

func (e Error) Type() Type      { return ERROR }
func (e Error) Inspect() string { return fmt.Sprintf("Error: %s", e) }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (fn *Function) Type() Type { return FUNCTION }
func (fn *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fn.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(fn.Body.String())
	out.WriteString("\n}")

	return out.String()
}

type BuiltInFunction func(args ...Object) Object

func (bf BuiltInFunction) Type() Type      { return BUILTIN }
func (bf BuiltInFunction) Inspect() string { return "builtin function" }

type Array struct {
	Elements []Object
}

func (a *Array) Type() Type { return ARRAY }
func (a *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

type Hash struct {
	Pairs map[String]Object
}

func (h *Hash) Type() Type { return HASH }
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for k, v := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", k.Inspect(), v.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}

type Exp struct{ ast.Expression }

type Quote struct{ ast.Node }

func (q *Quote) Type() Type { return QUOTE }
func (q *Quote) Inspect() string {
	return "QUOTE(" + q.Node.String() + ")"
}

type Macro struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (m *Macro) Type() Type { return FUNCTION }
func (m *Macro) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range m.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("macro")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(m.Body.String())
	out.WriteString("\n}")

	return out.String()
}
