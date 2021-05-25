package eval

import (
	"github.com/dikaeinstein/monkey/object"
)

var builtins = map[string]object.BuiltInFunction{
	"len":   object.GetBuiltinByName("len"),
	"first": object.GetBuiltinByName("first"),
	"rest":  object.GetBuiltinByName("rest"),
	"push":  object.GetBuiltinByName("push"),
	"puts":  object.GetBuiltinByName("puts"),
}
