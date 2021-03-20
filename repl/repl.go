package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/dikaeinstein/monkey/lexer"
	"github.com/dikaeinstein/monkey/parser"
)

const prompt = ">> "
const allowedNumOfErrors = 0

const MONKEY_FACE = `             __,__
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

		_, err := io.WriteString(out, program.String())
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
