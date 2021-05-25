package vm

import (
	"github.com/dikaeinstein/monkey/code"
	"github.com/dikaeinstein/monkey/object"
)

type Frame struct {
	cl          *object.Closure
	ip          int
	basePointer uint
}

func NewFrame(cl *object.Closure, basePointer uint) *Frame {
	return &Frame{cl: cl, ip: -1, basePointer: basePointer}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
