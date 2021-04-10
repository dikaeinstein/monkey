package eval

import (
	"github.com/dikaeinstein/monkey/ast"
	"github.com/dikaeinstein/monkey/object"
)

func DefineMacros(p *ast.Program, env *object.Environment) {
	n := 0
	for _, stmt := range p.Statements {
		// filter in place
		if !isMacroDefinition(stmt) {
			p.Statements[n] = stmt
			n++
		} else {
			addMacro(stmt, env)
		}
	}

	// reslice up to non-macro selected statements
	// to remove macros from the p.Statements slice
	p.Statements = p.Statements[:n]
}

func isMacroDefinition(node ast.Statement) bool {
	letStatement, ok := node.(*ast.LetStatement)
	if !ok {
		return false
	}
	_, ok = letStatement.Value.(*ast.MacroLiteral)

	return ok
}

func addMacro(stmt ast.Statement, env *object.Environment) {
	letStmt := stmt.(*ast.LetStatement)
	macroLit := letStmt.Value.(*ast.MacroLiteral)

	macro := &object.Macro{
		Parameters: macroLit.Parameters,
		Body:       macroLit.Body,
		Env:        env,
	}

	env.Set(letStmt.Name.Value, macro)
}

func ExpandMacros(p *ast.Program, env *object.Environment) ast.Node {
	return ast.Modify(p, func(node ast.Node) ast.Node {
		callExpression, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}

		macro, ok := isMacroCall(callExpression, env)
		if !ok {
			return node
		}

		args := quoteArgs(callExpression)
		evalEnv := extendMacroEnv(macro, args)
		evaluated := Eval(macro.Body, evalEnv)

		quote, ok := evaluated.(*object.Quote)
		if !ok {
			panic("we only support returning AST-nodes from macros")
		}
		return quote.Node
	})
}

func isMacroCall(exp *ast.CallExpression, env *object.Environment) (*object.Macro, bool) {
	ident, ok := exp.Function.(*ast.Identifier)
	if !ok {
		return nil, false
	}

	obj, ok := env.Get(ident.Value)
	if !ok {
		return nil, false
	}

	macro, ok := obj.(*object.Macro)
	return macro, ok
}

func quoteArgs(exp *ast.CallExpression) []*object.Quote {
	args := []*object.Quote{}

	for _, a := range exp.Arguments {
		args = append(args, &object.Quote{Node: a})
	}

	return args
}

func extendMacroEnv(macro *object.Macro, args []*object.Quote) *object.Environment {
	extended := macro.Env.AddFrame()

	for paramIdx, param := range macro.Parameters {
		extended.Set(param.Value, args[paramIdx])
	}

	return extended
}
