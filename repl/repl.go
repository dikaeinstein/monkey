package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/dikaeinstein/monkey/compile"
	"github.com/dikaeinstein/monkey/eval"
	"github.com/dikaeinstein/monkey/lexer"
	"github.com/dikaeinstein/monkey/object"
	"github.com/dikaeinstein/monkey/parser"
	"github.com/dikaeinstein/monkey/vm"
)

const prompt = ">> "
const allowedNumOfErrors = 0

// MonkeyFace is the face of our lovely mascot
const MonkeyFace = `             __,__
     .--. .-"     "-. .--.
    / .. \/ .-. .-. \/ .. \
   | |  '| /   Y   \ |'  | |
   | \   \ \ 0 | 0 / /   / |
   \ '- ,\.-"""""""-./, -' /
    ''-' /_   ^ ^   _\ '-''
        |  \._   _./  |
        \   \ '~' /   /
         '._ '-=-' _.'
            '-----'
`

// Start starts the REPL with the given io.Reader and io.Writer
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	// env := object.NewEnvironment()
	macroEnv := object.NewEnvironment()

	for {
		fmt.Print(prompt)
		if scanned := scanner.Scan(); !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != allowedNumOfErrors {
			printParserErrors(out, p.Errors())
			continue
		}

		eval.DefineMacros(program, macroEnv)
		expanded := eval.ExpandMacros(program, macroEnv)

		// TODO
		// evaluated := eval.Eval(expanded, env)
		// if evaluated != nil {
		// 	_, err := io.WriteString(out, evaluated.Inspect())
		// 	must(err)
		// 	_, err = io.WriteString(out, "\n")
		// 	must(err)
		// }

		compiler := compile.New()
		err := compiler.Compile(expanded)
		if err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		machine := vm.New(compiler.Bytecode())
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		machine.StackTop()
		stackTop := machine.StackTop()
		_, err = io.WriteString(out, stackTop.Inspect())
		must(err)
		_, err = io.WriteString(out, "\n")
		must(err)
	}
}

func printParserErrors(out io.Writer, errors []string) {
	_, err := io.WriteString(out, "Woops! We ran into some monkey business here!\n")
	must(err)
	_, err = io.WriteString(out, " parser errors:\n")
	must(err)

	for _, msg := range errors {
		_, err := io.WriteString(out, "\t"+msg+"\n")
		must(err)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
