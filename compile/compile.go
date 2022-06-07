package compile

import (
	"fmt"
	"sort"

	"github.com/dikaeinstein/monkey/ast"
	"github.com/dikaeinstein/monkey/code"
	"github.com/dikaeinstein/monkey/object"
	"github.com/dikaeinstein/monkey/token"
)

type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions code.Instructions

	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

// Compiler wraps the bytecode instructions and constants pool.
type Compiler struct {
	constants   []object.Object
	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int
}

// New returns a new instance of the Compiler
func New(symbolTable *SymbolTable, constants []object.Object) *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	return &Compiler{
		constants:   constants,
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
	}
}

// NewCompilerWithBuiltins returns a new instance of the Compiler,
// with builtin functions defined.
func NewCompilerWithBuiltins(constants []object.Object) *Compiler {
	symbolTable := NewSymbolTable()

	for i, def := range object.Builtins() {
		symbolTable.DefineBuiltin(i, def.Name)
	}

	return New(symbolTable, constants)
}

// Compile compiles the AST into Bytecode. It fills the compiler instructions
// and constant pool with compiled bytecode instructions and evaluated
// constants.
//gocyclo:ignore
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
		c.emit(code.OpPop)
	case *ast.InfixExpression:
		err := c.compileInfixExpression(node)
		if err != nil {
			return err
		}
	case *ast.PrefixExpression:
		err := c.compilePrefixExpression(node)
		if err != nil {
			return err
		}
	case *ast.IntegerLiteral:
		integer := object.Integer(node.Value)
		c.emit(code.OpConstant, c.addConstant(integer))
	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}
	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}
	case *ast.IfExpression:
		err := c.compileIfExpression(node)
		if err != nil {
			return err
		}
	case *ast.Identifier:
		sym, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable: %s", node.Value)
		}
		c.loadSymbol(sym)
	case *ast.LetStatement:
		err := c.compileLetStatement(node)
		if err != nil {
			return err
		}
	case *ast.StringLiteral:
		str := object.String(node.Value)
		c.emit(code.OpConstant, c.addConstant(str))
	case *ast.ArrayLiteral:
		for _, elem := range node.Elements {
			err := c.Compile(elem)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpArray, len(node.Elements))
	case *ast.HashLiteral:
		err := c.compileHashLiteral(node)
		if err != nil {
			return err
		}
	case *ast.IndexExpression:
		err := c.compileIndexExpression(node)
		if err != nil {
			return err
		}
	case *ast.FunctionLiteral:
		err := c.compileFunctionLiteral(node)
		if err != nil {
			return err
		}
	case *ast.ReturnStatement:
		err := c.Compile(node.ReturnValue)
		if err != nil {
			return err
		}

		c.emit(code.OpReturnValue)
	case *ast.CallExpression:
		err := c.compileCallExpression(node)
		if err != nil {
			return err
		}
	}

	return nil
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
	SymbolTable  *SymbolTable
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) compileInfixExpression(node *ast.InfixExpression) error {
	// invert the operands for lessThan operator so the compiler can use
	// the greaterThan operator
	if node.Operator == string(token.LT) {
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		err = c.Compile(node.Left)
		if err != nil {
			return err
		}

		c.emit(code.OpGreaterThan)
		return nil
	}

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
	case string(token.MINUS):
		c.emit(code.OpSub)
	case string(token.ASTERISK):
		c.emit(code.OpMul)
	case string(token.SLASH):
		c.emit(code.OpDiv)
	case string(token.EQ):
		c.emit(code.OpEqual)
	case string(token.NotEQ):
		c.emit(code.OpNotEqual)
	case string(token.GT):
		c.emit(code.OpGreaterThan)
	default:
		return fmt.Errorf("unknown operator %s", node.Operator)
	}

	return nil
}

func (c *Compiler) compilePrefixExpression(node *ast.PrefixExpression) error {
	err := c.Compile(node.Right)
	if err != nil {
		return err
	}

	switch node.Operator {
	case string(token.BANG):
		c.emit(code.OpBang)
	case string(token.MINUS):
		c.emit(code.OpMinus)
	default:
		return fmt.Errorf("unknown operator %s", node.Operator)
	}

	return nil
}

func (c *Compiler) compileIfExpression(node *ast.IfExpression) error {
	err := c.Compile(node.Condition)
	if err != nil {
		return err
	}

	const bogus = 9999
	// Emit an `OpJumpNotTruthy` with a bogus value
	jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, bogus)
	err = c.Compile(node.Consequence)
	if err != nil {
		return err
	}

	if c.lastInstructionIs(code.OpPop) {
		c.removeLastPop()
	}

	// Emit an `OpJump` with a bogus value
	jumpPos := c.emit(code.OpJump, bogus)

	afterConsequencePos := len(c.currentInstructions())
	c.changeOperands(jumpNotTruthyPos, afterConsequencePos)

	if node.Alternative == nil {
		c.emit(code.OpNull)
	} else {
		afterConsequencePos := len(c.currentInstructions())
		c.changeOperands(jumpNotTruthyPos, afterConsequencePos)

		err = c.Compile(node.Alternative)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}
	}

	afterAlternativePos := len(c.currentInstructions())
	c.changeOperands(jumpPos, afterAlternativePos)

	return nil
}

func (c *Compiler) compileFunctionLiteral(node *ast.FunctionLiteral) error {
	c.enterScope()

	if node.Name != "" {
		c.symbolTable.DefineFunctionName(node.Name)
	}

	for _, p := range node.Parameters {
		c.symbolTable.Define(p.Value)
	}

	err := c.Compile(node.Body)
	if err != nil {
		return err
	}

	if c.lastInstructionIs(code.OpPop) {
		c.replaceLastPopWithReturn()
	}
	// when function body is empty, emit OpReturn
	if !c.lastInstructionIs(code.OpReturnValue) {
		c.emit(code.OpReturn)
	}

	freeSymbols := c.symbolTable.FreeSymbols
	numLocals := c.symbolTable.numDefinitions
	instructions := c.leaveScope()

	for _, s := range freeSymbols {
		c.loadSymbol(s)
	}

	compiledFn := &object.CompiledFunction{
		Instructions:  instructions,
		NumLocals:     numLocals,
		NumParameters: len(node.Parameters),
	}
	c.emit(code.OpClosure, c.addConstant(compiledFn), len(freeSymbols))

	return nil
}

func (c *Compiler) compileHashLiteral(node *ast.HashLiteral) error {
	keys := []ast.Expression{}
	for k := range node.Pairs {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].String() < keys[j].String()
	})
	for _, k := range keys {
		err := c.Compile(k)
		if err != nil {
			return err
		}
		err = c.Compile(node.Pairs[k])
		if err != nil {
			return err
		}
	}

	numBytePerPair := 2
	c.emit(code.OpHash, len(node.Pairs)*numBytePerPair)
	return nil
}

func (c *Compiler) compileCallExpression(node *ast.CallExpression) error {
	err := c.Compile(node.Function)
	if err != nil {
		return err
	}

	for _, arg := range node.Arguments {
		err := c.Compile(arg)
		if err != nil {
			return err
		}
	}

	c.emit(code.OpCall, len(node.Arguments))
	return nil
}

func (c *Compiler) compileLetStatement(node *ast.LetStatement) error {
	sym := c.symbolTable.Define(node.Name.Value)

	err := c.Compile(node.Value)
	if err != nil {
		return err
	}

	if sym.Scope == GlobalScope {
		c.emit(code.OpSetGlobal, sym.Index)
	} else {
		c.emit(code.OpSetLocal, sym.Index)
	}

	return nil
}

func (c *Compiler) compileIndexExpression(node *ast.IndexExpression) error {
	err := c.Compile(node.Left)
	if err != nil {
		return err
	}
	err = c.Compile(node.Index)
	if err != nil {
		return err
	}

	c.emit(code.OpIndex)
	return nil
}

func (c *Compiler) loadSymbol(sym Symbol) {
	switch sym.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, sym.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, sym.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, sym.Index)
	case FreeScope:
		c.emit(code.OpGetFree, sym.Index)
	case FunctionScope:
		c.emit(code.OpCurrentClosure)
	}
}

// currentInstructions return the instructions of the current scope
func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

// addConstant adds obj value to the compiler constant pool.
// It returns the current index in the pool.
func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// emit generates a bytecode instruction and saves it to the compiler
// list of instructions.
func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	pos := c.addInstruction(code.Make(op, operands...))
	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	c.scopes[c.scopeIndex].previousInstruction = c.scopes[c.scopeIndex].lastInstruction
	c.scopes[c.scopeIndex].lastInstruction = EmittedInstruction{Opcode: op, Position: pos}
}

// addInstruction adds the given instruction to the compilers list of
// instructions. It returns the starting position of the instruction.
func (c *Compiler) addInstruction(ins []byte) int {
	posNewIns := len(c.currentInstructions())

	updatedInstructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return posNewIns
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	prev := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()

	c.scopes[c.scopeIndex].instructions = old[:last.Position]
	c.scopes[c.scopeIndex].lastInstruction = prev
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	for i := 0; i < len(newInstruction); i++ {
		c.currentInstructions()[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

// changeOperands replaces the operand for an Opcode using its position
// in the compiler instruction list.
func (c *Compiler) changeOperands(opPos, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) enterScope() {
	newScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	c.scopes = append(c.scopes, newScope)
	c.scopeIndex++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.symbolTable = c.symbolTable.parent

	return instructions
}
