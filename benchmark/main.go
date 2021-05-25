package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/dikaeinstein/monkey/compile"
	"github.com/dikaeinstein/monkey/eval"
	"github.com/dikaeinstein/monkey/lexer"
	"github.com/dikaeinstein/monkey/object"
	"github.com/dikaeinstein/monkey/parser"
	"github.com/dikaeinstein/monkey/vm"
)

func main() {
	engine := flag.String("engine", "vm", "use 'vm' or 'eval'")
	flag.Parse()

	var input = `
	let fibonacci = fn(x) {
		if (x == 0) {
			0
		} else {
			if (x == 1) {
				return 1;
			} else {
				fibonacci(x - 1) + fibonacci(x - 2);
			}
		}
	};

	fibonacci(35);
	`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Println("parser errors:", p.Errors())
		return
	}

	var duration time.Duration
	var result object.Object

	if *engine == "vm" {
		compiler := compile.NewCompilerWithBuiltins([]object.Object{})
		err := compiler.Compile(program)
		if err != nil {
			fmt.Printf("compiler error: %s", err)
			return
		}

		machine := vm.New(compiler.Bytecode())

		start := time.Now()

		err = machine.Run()
		if err != nil {
			fmt.Printf("vm error: %s", err)
			return
		}

		duration = time.Since(start)
		result = machine.LastPoppedStackElem()
	} else {
		env := object.NewEnvironment()
		start := time.Now()
		result = eval.Eval(program, env)
		duration = time.Since(start)
	}

	fmt.Printf(
		"engine=%s, result=%s, duration=%s\n",
		*engine,
		result.Inspect(),
		duration)
}
