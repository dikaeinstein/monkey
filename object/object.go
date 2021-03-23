package object

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/dikaeinstein/monkey/ast"
)

type ObjectType string

// Object represents values in the host language(GO).
// An object can either be a primitive or reference value.
type Object interface {
	Type() ObjectType
	Inspect() string
}

const (
	ARRAY    ObjectType = "ARRAY"
	BOOLEAN  ObjectType = "BOOLEAN"
	BUILTIN  ObjectType = "BUILTIN"
	ERROR    ObjectType = "ERROR"
	FUNCTION ObjectType = "FUNCTION"
	HASH     ObjectType = "HASH"
	INTEGER  ObjectType = "INTEGER"
	NULL     ObjectType = "NULL"
	STRING   ObjectType = "STRING"
)

type Integer int64

func (i Integer) Type() ObjectType { return INTEGER }
func (i Integer) Inspect() string  { return fmt.Sprint(i) }

type String string

func (s String) Type() ObjectType { return STRING }
func (s String) Inspect() string  { return string(s) }

type Boolean bool

func (b Boolean) Type() ObjectType { return BOOLEAN }
func (b Boolean) Inspect() string  { return fmt.Sprint(b) }

type Null struct{}

func (n Null) Type() ObjectType { return NULL }
func (n Null) Inspect() string  { return "null" }

type Error string

func (e Error) Type() ObjectType { return ERROR }
func (e Error) Inspect() string  { return fmt.Sprintf("Error: %s", e) }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (fn *Function) Type() ObjectType { return FUNCTION }
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

func (bf BuiltInFunction) Type() ObjectType { return BUILTIN }
func (bf BuiltInFunction) Inspect() string  { return "builtin function" }

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY }
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

func (h *Hash) Type() ObjectType { return HASH }
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
