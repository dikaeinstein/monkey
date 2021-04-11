package compile

import (
	"fmt"

	"github.com/dikaeinstein/monkey/ast"
	"github.com/dikaeinstein/monkey/code"
	"github.com/dikaeinstein/monkey/object"
	"github.com/dikaeinstein/monkey/token"
)

// Compiler wraps the bytecode instructions and constant pool.
type Compiler struct {
	instructions code.Instructions
	constants    []object.Object
}

// New returns a new initialized instance of the Compiler
func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []object.Object{},
	}
}

// Compile compiles the AST into Bytecode. It fills the compiler instructions
// and constant pool with compiled bytecode instructions and evaluated
// constants.
func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
	case *ast.InfixExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case string(token.PLUS):
			c.emit(code.OpAdd)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}
	case *ast.IntegerLiteral:
		integer := object.Integer(node.Value)
		c.emit(code.OpConstant, c.addConstant(integer))
	}

	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}

// addConstant adds obj value to compiler constant pool.
// It returns the current index in the pool.
func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// emit generates a bytecode instruction and saves it to the compiler
// list of instructions. It also writes constants to the constant pool
// if found.
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	return c.addInstruction(code.Make(op, operands...))
}

// addInstruction adds the given instruction to the compilers list of
// instructions. It returns the starting position of the instruction.
func (c *Compiler) addInstruction(ins []byte) int {
	posNewIns := len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewIns
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}
