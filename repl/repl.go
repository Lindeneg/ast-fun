package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/lindeneg/monkey/evaluator"
	"github.com/lindeneg/monkey/lexer"
	"github.com/lindeneg/monkey/object"
	"github.com/lindeneg/monkey/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.NewLexer(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			PrintParserErrors(out, p.Errors())
			continue
		}
		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func PrintParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
