package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/dikaeinstein/monkey/lexer"
	"github.com/dikaeinstein/monkey/token"
)

const prompt = ">> "

// Start starts the REPL with the given reader and writer
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Print(prompt)
		if scanned := scanner.Scan(); !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			fmt.Printf("%+v\n", tok)
		}
	}
}
