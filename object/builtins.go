package object

import (
	"fmt"
)

type NamedBuiltinFunction struct {
	Name    string
	Builtin BuiltInFunction
}

var builtins = []NamedBuiltinFunction{
	{
		Name: "len",
		Builtin: func(args ...Object) Object {
			const allowedNumOfArgs = 1
			err := checkArgsLen(allowedNumOfArgs, args)
			if IsError(err) {
				return err
			}

			switch arg := args[0].(type) {
			case String:
				return Integer(len(arg))
			case *Array:
				return Integer(len(arg.Elements))
			default:
				return newError("argument to `len` not supported, got %s", arg.Type())
			}
		},
	},
	{
		Name: "first",
		Builtin: func(args ...Object) Object {
			const allowedNumOfArgs = 1
			err := checkArgsLen(allowedNumOfArgs, args)
			if IsError(err) {
				return err
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to `first` must be ARRAY, got %s",
					args[0].Type())
			}
			if len(arr.Elements) < 1 {
				return NullValue()
			}

			return arr.Elements[0]
		},
	},
	{
		Name: "last",
		Builtin: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1",
					len(args))
			}
			if args[0].Type() != ARRAY {
				return newError("argument to `last` must be ARRAY, got %s",
					args[0].Type())
			}

			arr := args[0].(*Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}

			return nil
		},
	},
	{
		Name: "rest",
		Builtin: func(args ...Object) Object {
			const allowedNumOfArgs = 1
			err := checkArgsLen(allowedNumOfArgs, args)
			if IsError(err) {
				return err
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to `rest` must be ARRAY, got %s",
					args[0].Type())
			}
			if len(arr.Elements) < 1 {
				return NullValue()
			}

			size := len(arr.Elements)
			newElements := make([]Object, size-1)
			copy(newElements, arr.Elements[1:size])
			return &Array{Elements: newElements}
		},
	},
	{
		Name: "push",
		Builtin: func(args ...Object) Object {
			const allowedNumOfArgs = 2
			err := checkArgsLen(allowedNumOfArgs, args)
			if IsError(err) {
				return err
			}

			arr, ok := args[0].(*Array)
			if !ok {
				return newError("argument to `push` must be ARRAY, got %s",
					args[0].Type())
			}

			size := len(arr.Elements)
			newElements := make([]Object, size+1)
			copy(newElements, arr.Elements)
			newElements[size] = args[1]

			return &Array{Elements: newElements}
		},
	},
	{
		Name: "puts",
		Builtin: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}

			return nil
		},
	},
}

func Builtins() []NamedBuiltinFunction {
	return builtins
}

func GetBuiltinByName(name string) BuiltInFunction {
	for _, def := range builtins {
		if def.Name == name {
			return def.Builtin
		}
	}
	return nil
}

func checkArgsLen(expectedLen int, args []Object) Object {
	if len(args) != expectedLen {
		return newError("wrong number of arguments. got=%d, want=%d",
			len(args), expectedLen)
	}
	return nil
}

func newError(format string, a ...interface{}) Error {
	return Error(fmt.Sprintf(format, a...))
}

func IsError(obj Object) bool {
	if obj != nil {
		return obj.Type() == ERROR
	}
	return false
}
