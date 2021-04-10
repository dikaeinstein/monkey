package eval

import (
	"fmt"

	"github.com/dikaeinstein/monkey/ast"
	"github.com/dikaeinstein/monkey/object"
	"github.com/dikaeinstein/monkey/token"
)

// The one and only null value
var defaultNull = &object.Null{}

// Eval evaluates walks the code by walking the parsed AST
//gocyclo:ignore
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
	case *ast.StringLiteral:
		return object.String(node.Value)
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
	case *ast.ArrayLiteral:
		return evalArrayLiteral(node, env)
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)
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
	case left.Type() == object.STRING && right.Type() == object.STRING:
		return evalStringInfixExpression(node.Operator, left, right)
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
	case string(token.NotEQ):
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
	case string(token.NotEQ):
		return object.Boolean(lVal != rVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	lVal := left.(object.String)
	rVal := right.(object.String)

	if operator != string(token.PLUS) {
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}

	return lVal + rVal
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

// evalIdentifier resolve names in this order: (local, enclosing, global, builtin)
func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if val, ok := builtins[node.Value]; ok {
		return val
	}

	return newError("identifier not found: " + node.Value)
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
	if node.Function.TokenLiteral() == "quote" {
		return quote(node.Arguments[0], env)
	}

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
	switch function := fn.(type) {
	case *object.Function:
		// allocate a new frame for function on top of the stack
		function.Env.AddFrame()
		// bind arguments to parameters in the function stack frame a.k.a scope
		for i, arg := range args {
			ident := function.Parameters[i]
			function.Env.Set(ident.Value, arg)
		}

		return evalStatements(function.Body.Statements, function.Env)
	case object.BuiltInFunction:
		// use function already defined with host lang(Go)
		return function(args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func evalArrayLiteral(node *ast.ArrayLiteral, env *object.Environment) object.Object {
	elements := evalExpressions(node.Elements, env)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}
	return &object.Array{Elements: elements}
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.INTEGER:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(left, index object.Object) object.Object {
	array := left.(*object.Array)
	idx := index.(object.Integer)
	max := int64(len(array.Elements) - 1)

	if idx < 0 || int64(idx) > max {
		return null()
	}

	return array.Elements[idx]
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	h := &object.Hash{Pairs: make(map[object.String]object.Object)}

	for k, v := range node.Pairs {
		kk := evalHashKey(Eval(k, env))
		if isError(kk) {
			return kk
		}

		key := kk.(object.String)
		value := Eval(v, env)
		if isError(value) {
			return value
		}

		h.Pairs[key] = value
	}

	return h
}

func evalHashKey(key object.Object) object.Object {
	switch key := key.(type) {
	case object.String:
		return key
	case object.Integer:
		return object.String(key.Inspect())
	case object.Boolean:
		return object.String(key.Inspect())
	default:
		return newError("unusable as hash key: %s", key.Type())
	}
}

func evalHashIndexExpression(left, index object.Object) object.Object {
	hash := left.(*object.Hash)
	kk := evalHashKey(index)
	if isError(kk) {
		return kk
	}

	key := kk.(object.String)
	val, ok := hash.Pairs[key]
	if !ok {
		return null()
	}

	return val
}

func quote(node ast.Node, env *object.Environment) object.Object {
	node = evalUnquoteCalls(node, env)
	return &object.Quote{Node: node}
}

func evalUnquoteCalls(node ast.Node, env *object.Environment) ast.Node {
	return ast.Modify(node, func(node ast.Node) ast.Node {
		if !isUnquotedCall(node) {
			return node
		}

		call, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}

		if len(call.Arguments) != 1 {
			return node
		}

		return convertObjectToASTNode(Eval(call.Arguments[0], env))
	})
}

func isUnquotedCall(node ast.Node) bool {
	exp, ok := node.(*ast.CallExpression)
	if !ok {
		return false
	}

	return exp.Function.TokenLiteral() == "unquote"
}

func convertObjectToASTNode(obj object.Object) ast.Node {
	switch obj := obj.(type) {
	case object.Integer:
		t := token.Token{
			Type:    token.INT,
			Literal: fmt.Sprintf("%d", obj),
		}
		return &ast.IntegerLiteral{Token: t, Value: int64(obj)}
	case object.Boolean:
		var t token.Token
		if obj {
			t = token.Token{Type: token.TRUE, Literal: "true"}
		} else {
			t = token.Token{Type: token.FALSE, Literal: "false"}
		}
		return &ast.Boolean{Token: t, Value: bool(obj)}
	case *object.Quote:
		return obj.Node
	default:
		return nil
	}
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
