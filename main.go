package main

import (
	"fmt"
	"log"
	"os"
	"os/user"

	"github.com/lindeneg/monkey/evaluator"
	"github.com/lindeneg/monkey/lexer"
	"github.com/lindeneg/monkey/object"
	"github.com/lindeneg/monkey/parser"
	"github.com/lindeneg/monkey/repl"
)

func main() {
	if len(os.Args) > 1 {
		data, err := os.ReadFile(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		env := object.NewEnvironment()
		l := lexer.NewLexer(string(data))
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) > 0 {
			repl.PrintParserErrors(os.Stdout, p.Errors())
		}
		result := evaluator.Eval(program, env)
		if result != nil && result.Type() == object.ERROR_OBJ {
			fmt.Println(result.Inspect())
		}
	} else {
		u, err := user.Current()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Hello %s! This is the Monkey programming language!\n",
			u.Username)
		fmt.Printf("Feel free to type in commands\n")
		repl.Start(os.Stdin, os.Stdout)
	}
}
