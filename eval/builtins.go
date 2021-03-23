package eval

import (
	"fmt"

	"github.com/dikaeinstein/monkey/object"
)

var builtins = map[string]object.BuiltInFunction{
	"len": func(args ...object.Object) object.Object {
		const allowedNumOfArgs = 1
		err := checkArgsLen(allowedNumOfArgs, args)
		if isError(err) {
			return err
		}

		switch arg := args[0].(type) {
		case object.String:
			return object.Integer(len(arg))
		case *object.Array:
			return object.Integer(len(arg.Elements))
		default:
			return newError("argument to `len` not supported, got %s", arg.Type())
		}
	},
	"first": func(args ...object.Object) object.Object {
		const allowedNumOfArgs = 1
		err := checkArgsLen(allowedNumOfArgs, args)
		if isError(err) {
			return err
		}

		arr, ok := args[0].(*object.Array)
		if !ok {
			return newError("argument to `first` must be ARRAY, got %s",
				args[0].Type())
		}
		if len(arr.Elements) < 1 {
			return null()
		}

		return arr.Elements[0]
	},
	"rest": func(args ...object.Object) object.Object {
		const allowedNumOfArgs = 1
		err := checkArgsLen(allowedNumOfArgs, args)
		if isError(err) {
			return err
		}

		arr, ok := args[0].(*object.Array)
		if !ok {
			return newError("argument to `rest` must be ARRAY, got %s",
				args[0].Type())
		}
		if len(arr.Elements) < 1 {
			return null()
		}

		size := len(arr.Elements)
		newElements := make([]object.Object, size-1)
		copy(newElements, arr.Elements[1:size])
		return &object.Array{Elements: newElements}
	},
	"push": func(args ...object.Object) object.Object {
		const allowedNumOfArgs = 2
		err := checkArgsLen(allowedNumOfArgs, args)
		if isError(err) {
			return err
		}

		arr, ok := args[0].(*object.Array)
		if !ok {
			return newError("argument to `push` must be ARRAY, got %s",
				args[0].Type())
		}

		size := len(arr.Elements)
		newElements := make([]object.Object, size+1)
		copy(newElements, arr.Elements)
		newElements[size] = args[1]

		return &object.Array{Elements: newElements}
	},
	"puts": func(args ...object.Object) object.Object {
		for _, arg := range args {
			fmt.Println(arg.Inspect())
		}

		return null()
	},
}

func checkArgsLen(expectedLen int, args []object.Object) object.Object {
	if len(args) != expectedLen {
		return newError("wrong number of arguments. got=%d, want=%d",
			len(args), expectedLen)
	}
	return nil
}
