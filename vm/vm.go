package vm

import (
	"fmt"

	"github.com/dikaeinstein/monkey/code"
	"github.com/dikaeinstein/monkey/compile"
	"github.com/dikaeinstein/monkey/object"
)

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1024

type VM struct {
	constants []object.Object

	frames      []*Frame
	framesIndex int

	globals []object.Object

	stack []object.Object
	sp    uint // Always points to the next value. Top of stack is stack[sp-1]
}

func NewWithGlobalsStore(bytecode *compile.Bytecode, globals []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = globals

	return vm
}

func New(bytecode *compile.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	globals := make([]object.Object, GlobalsSize)

	return &VM{
		constants: bytecode.Constants,

		frames:      frames,
		framesIndex: 1,

		globals: globals,

		stack: make([]object.Object, StackSize),
		sp:    0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

//gocyclo:ignore
// Run fetches, decodes and executes the bytecode instructions
func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += code.OperandWidth2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return nil
			}
		case code.OpPop:
			vm.pop()
		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}
		case code.OpTrue:
			err := vm.push(object.Boolean(true))
			if err != nil {
				return err
			}
		case code.OpFalse:
			err := vm.push(object.Boolean(false))
			if err != nil {
				return err
			}
		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}
		case code.OpBang, code.OpMinus:
			err := vm.executePrefixExpression(op)
			if err != nil {
				return err
			}
		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			// decrement pos, so the FDE loop does its work to set the ip
			// correctly in the next cycle
			vm.currentFrame().ip = pos - 1
		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += code.OperandWidth2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}
		case code.OpNull:
			err := vm.push(object.NullValue())
			if err != nil {
				return err
			}
		case code.OpSetGlobal:
			symbolIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += code.OperandWidth2

			vm.globals[symbolIndex] = vm.pop()
		case code.OpGetGlobal:
			symbolIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += code.OperandWidth2

			err := vm.push(vm.globals[symbolIndex])
			if err != nil {
				return err
			}
		case code.OpArray:
			numOfElements := uint(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += code.OperandWidth2

			array := vm.buildArray(vm.sp-numOfElements, vm.sp)
			vm.sp -= numOfElements

			err := vm.push(array)
			if err != nil {
				return err
			}
		case code.OpHash:
			numOfElements := uint(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-numOfElements, vm.sp)
			if err != nil {
				return err
			}

			vm.sp -= numOfElements
			err = vm.push(hash)

			if err != nil {
				return err
			}
		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}
		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip++

			err := vm.executeCall(numArgs)
			if err != nil {
				return err
			}
		case code.OpReturnValue:
			returnValue := vm.pop()

			frame := vm.popFrame()        // leave fn stack frame
			vm.sp = frame.basePointer - 1 // move stack pointer back to point before compiledFn

			err := vm.push(returnValue)
			if err != nil {
				return err
			}
		case code.OpReturn:
			frame := vm.popFrame()        // leave fn stack frame
			vm.sp = frame.basePointer - 1 // move stack pointer back to point before compiledFn

			err := vm.push(object.NullValue())
			if err != nil {
				return err
			}
		case code.OpSetLocal:
			localIdx := code.ReadUint8(ins[vm.currentFrame().ip+1:])
			vm.currentFrame().ip += code.OperandWidth1

			frame := vm.currentFrame()
			vm.stack[frame.basePointer+uint(localIdx)] = vm.pop()
		case code.OpGetLocal:
			localIdx := code.ReadUint8(ins[vm.currentFrame().ip+1:])
			vm.currentFrame().ip += code.OperandWidth1

			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.basePointer+uint(localIdx)])
			if err != nil {
				return err
			}
		case code.OpGetBuiltin:
			builtinFnIdx := code.ReadUint8(ins[vm.currentFrame().ip+1:])
			vm.currentFrame().ip += code.OperandWidth1

			builtins := object.Builtins()
			definition := builtins[builtinFnIdx]

			err := vm.push(definition.Builtin)
			if err != nil {
				return err
			}
		case code.OpClosure:
			constIndex := code.ReadUint16(ins[vm.currentFrame().ip+1:])
			numFree := code.ReadUint8(ins[vm.currentFrame().ip+3:])
			vm.currentFrame().ip += (code.OperandWidth2 + code.OperandWidth1)

			err := vm.pushClosure(constIndex, int(numFree))
			if err != nil {
				return err
			}
		case code.OpGetFree:
			freeIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip++

			currentClosure := vm.currentFrame().cl

			err := vm.push(currentClosure.Free[freeIndex])
			if err != nil {
				return err
			}
		case code.OpCurrentClosure:
			currentClosure := vm.currentFrame().cl

			err := vm.push(currentClosure)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

func (vm *VM) buildArray(startIndex, endIndex uint) object.Object {
	elements := make([]object.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex uint) (object.Object, error) {
	pairs := make(map[object.String]object.Object)

	for i := startIndex; i < endIndex; i += 2 {
		k := vm.stack[i]
		value := vm.stack[i+1]

		key, ok := object.IsHashable(k)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		pairs[key] = value
	}

	return &object.Hash{Pairs: pairs}, nil
}

func (vm *VM) push(obj object.Object) error {
	if vm.sp > StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = obj
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	obj := vm.stack[vm.sp-1]
	vm.sp--
	return obj
}

func (vm *VM) pushClosure(constIndex uint16, numFree int) error {
	constant := vm.constants[constIndex]
	fn, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[int(vm.sp)-numFree+i]
	}
	vm.sp -= uint(numFree)

	return vm.push(&object.Closure{Fn: fn, Free: free})
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER && rightType == object.INTEGER:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.STRING && rightType == object.STRING:
		return vm.executeBinaryStringOperation(op, left, right)
	}

	return fmt.Errorf("unsupported types for binary operation: %s %s",
		leftType, rightType)
}

func (vm *VM) executeBinaryIntegerOperation(
	op code.Opcode,
	left, right object.Object,
) error {
	leftValue := left.(object.Integer)
	rightValue := right.(object.Integer)

	var result object.Integer
	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(result)
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftValue := left.(object.String)
	rightValue := right.(object.String)

	return vm.push(leftValue + rightValue)
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER && right.Type() == object.INTEGER {
		return vm.executeIntegerComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(object.Boolean(right == left))
	case code.OpNotEqual:
		return vm.push(object.Boolean(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)",
			op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(
	op code.Opcode,
	left, right object.Object,
) error {
	leftValue := left.(object.Integer)
	rightValue := right.(object.Integer)

	switch op {
	case code.OpEqual:
		return vm.push(object.Boolean(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(object.Boolean(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(object.Boolean(leftValue > rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executePrefixExpression(op code.Opcode) error {
	switch op {
	case code.OpBang:
		return vm.executeBangOperator()
	case code.OpMinus:
		operand := vm.pop()
		switch operand := operand.(type) {
		case object.Integer:
			return vm.push(-operand)
		default:
			return fmt.Errorf("unsupported type for prefix expression: %s",
				operand.Type())
		}
	default:
		return fmt.Errorf("unsupported opCode for prefix expression: %d", op)
	}
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case object.Boolean(true):
		return vm.push(object.Boolean(true))
	case object.Boolean(false):
		return vm.push(object.Boolean(false))
	case object.NullValue():
		return vm.push(object.Boolean(true))
	default:
		return vm.push(object.Boolean(false))
	}
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY && index.Type() == object.INTEGER:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH:
		return vm.executeHashIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(object.Integer)
	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || int64(i) > max {
		return vm.push(object.NullValue())
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)

	key, ok := object.IsHashable(index)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	value, ok := hashObject.Pairs[key]
	if !ok {
		return vm.push(object.NullValue())
	}

	return vm.push(value)
}

func (vm *VM) executeCall(numArgs uint8) error {
	callee := vm.stack[vm.sp-1-uint(numArgs)]

	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case object.BuiltInFunction:
		return vm.callBuiltinFn(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function and non-built-in")
	}
}

func (vm *VM) callClosure(cl *object.Closure, numArgs uint8) error {
	if numArgs != uint8(cl.Fn.NumParameters) {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			cl.Fn.NumParameters, numArgs)
	}

	frame := NewFrame(cl, vm.sp-uint(numArgs))
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + uint(cl.Fn.NumLocals)

	return nil
}

func (vm *VM) callBuiltinFn(fn object.BuiltInFunction, numArgs uint8) error {
	args := vm.stack[vm.sp-uint(numArgs) : vm.sp]

	result := fn(args...)

	var err error
	if result != nil {
		err = vm.push(result)
	} else {
		err = vm.push(object.NullValue())
	}

	return err
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case object.Boolean:
		return bool(obj)
	case *object.Null:
		return false
	// case object.Integer:
	// 	return int64(obj) != 0
	default:
		return true
	}
}
