package vm

import (
	"fmt"

	"github.com/dikaeinstein/monkey/code"
	"github.com/dikaeinstein/monkey/compile"
	"github.com/dikaeinstein/monkey/object"
)

const StackSize = 2048

type VM struct {
	constants    []object.Object
	instructions code.Instructions

	stack []object.Object
	sp    uint // Always points to the next value. Top of stack is stack[sp-1]
}

func New(bytecode *compile.Bytecode) *VM {
	return &VM{
		constants:    bytecode.Constants,
		instructions: bytecode.Instructions,
		stack:        make([]object.Object, StackSize),
		sp:           0,
	}
}

func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

func (vm *VM) Run() error {
	for ip := 0; ip < len(vm.instructions); ip++ {
		op := code.Opcode(vm.instructions[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(vm.instructions[ip+1:])
			ip += code.OperandWidth2

			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return nil
			}
		case code.OpAdd:
			right := vm.pop()
			left := vm.pop()
			lVal := left.(object.Integer)
			rVal := right.(object.Integer)

			result := lVal + rVal
			err := vm.push(result)
			if err != nil {
				return nil
			}
		}
	}

	return nil
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
