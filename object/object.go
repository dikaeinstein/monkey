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
	BOOLEAN  ObjectType = "BOOLEAN"
	ERROR    ObjectType = "ERROR"
	FUNCTION ObjectType = "FUNCTION"
	INTEGER  ObjectType = "INTEGER"
	NULL     ObjectType = "NULL"
)

type Integer int64

func (i Integer) Type() ObjectType { return INTEGER }
func (i Integer) Inspect() string  { return fmt.Sprint(i) }

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
