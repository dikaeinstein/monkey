package eval

import (
	"fmt"

	"github.com/dikaeinstein/monkey/ast"
	"github.com/dikaeinstein/monkey/object"
	"github.com/dikaeinstein/monkey/token"
)

// The one and only null value
var defaultNull = &object.Null{}

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	// Statements
	case *ast.Program:
		return evalStatements(node.Statements, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.BlockStatement:
		return evalStatements(node.Statements, env)
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		return nil
	case *ast.ReturnStatement:
		return Eval(node.ReturnValue, env)
	// Expressions
	case *ast.IntegerLiteral:
		return object.Integer(node.Value)
	case *ast.Boolean:
		return object.Boolean(node.Value)
	case *ast.PrefixExpression:
		return evalPrefixExpression(node, env)
	case *ast.InfixExpression:
		return evalInfixExpression(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		return evalFunction(node, env)
	case *ast.CallExpression:
		return evalCallExpression(node, env)
	default:
		return nil
	}
}

func evalStatements(statements []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, stmt := range statements {
		if _, ok := stmt.(*ast.ReturnStatement); ok {
			return Eval(stmt, env)
		}
		result = Eval(stmt, env)
		if isError(result) {
			return result
		}
	}

	return result
}

func evalPrefixExpression(node *ast.PrefixExpression, env *object.Environment) object.Object {
	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}

	switch node.Operator {
	case string(token.BANG):
		return evalBangOperatorExpression(right)
	case string(token.MINUS):
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", node.Operator, right.Type())
	}
}

func evalInfixExpression(node *ast.InfixExpression, env *object.Environment) object.Object {
	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}
	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}

	switch {
	case left.Type() == object.BOOLEAN && right.Type() == object.BOOLEAN:
		return evalBooleanInfixExpression(node.Operator, left, right)
	case left.Type() == object.INTEGER && right.Type() == object.INTEGER:
		return evalIntegerInfixExpression(node.Operator, left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), node.Operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), node.Operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	lVal := int64(left.(object.Integer))
	rVal := int64(right.(object.Integer))

	switch operator {
	case string(token.PLUS):
		return object.Integer(lVal + rVal)
	case string(token.MINUS):
		return object.Integer(lVal - rVal)
	case string(token.ASTERISK):
		return object.Integer(lVal * rVal)
	case string(token.SLASH):
		return object.Integer(lVal / rVal)
	case string(token.EQ):
		return object.Boolean(lVal == rVal)
	case string(token.NOT_EQ):
		return object.Boolean(lVal != rVal)
	case string(token.GT):
		return object.Boolean(lVal > rVal)
	case string(token.LT):
		return object.Boolean(lVal < rVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalBooleanInfixExpression(operator string, left, right object.Object) object.Object {
	lVal := bool(left.(object.Boolean))
	rVal := bool(right.(object.Boolean))

	switch operator {
	case string(token.EQ):
		return object.Boolean(lVal == rVal)
	case string(token.NOT_EQ):
		return object.Boolean(lVal != rVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right := right.(type) {
	case object.Boolean:
		return object.Boolean(!bool(right))
	case object.Null:
		return object.Boolean(true)
	default:
		return object.Boolean(false)
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	intVal, ok := right.(object.Integer)
	if !ok {
		return newError("unknown operator: -%s", right.Type())
	}
	return object.Integer(-int64(intVal))
}

// func evalBlockStatement(node *ast.BlockStatement) object.Object {
// 	return evalStatements(node.Statements)
// }

func evalIfExpression(node *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(node.Condition, env)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(node.Consequence, env)
	} else if node.Alternative != nil {
		return Eval(node.Alternative, env)
	} else {
		return null()
	}
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	val, ok := env.Get(node.Value)
	if !ok {
		return newError("identifier not found: " + node.Value)
	}
	return val
}

func evalFunction(node *ast.FunctionLiteral, env *object.Environment) object.Object {
	return &object.Function{
		Parameters: node.Parameters,
		Body:       node.Body,
		Env:        env,
	}
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

func evalCallExpression(node *ast.CallExpression, env *object.Environment) object.Object {
	fn := Eval(node.Function, env)
	if isError(fn) {
		return fn
	}

	args := evalExpressions(node.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	return applyFunction(fn, args)
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	function, ok := fn.(*object.Function)
	if !ok {
		return newError("not a function: %s", fn.Type())
	}

	// allocate a new frame for function on top of the stack
	function.Env.AddFrame()
	// bind arguments to parameters in the function stack frame a.k.a scope
	for i, arg := range args {
		ident := function.Parameters[i]
		function.Env.Set(ident.Value, arg)
	}

	return evalStatements(function.Body.Statements, function.Env)
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case object.Boolean:
		return bool(obj)
	case object.Null:
		return false
	case object.Integer:
		return int64(obj) != 0
	default:
		return true
	}
}

func newError(format string, a ...interface{}) object.Error {
	return object.Error(fmt.Sprintf(format, a...))
}

func null() *object.Null {
	return defaultNull
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR
	}
	return false
}