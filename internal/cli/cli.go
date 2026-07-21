package cli

import (
	"bufio"
	"fmt"
	"glox/pkg/interpreter"
	"glox/pkg/lexer"
	"glox/pkg/parser"
	"glox/pkg/token"
	"io"
	"os"
)

func StartRepl(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	interpreter := interpreter.New(out)

	for {
		fmt.Fprint(out, "> ")

		scanned := scanner.Scan()
		if !scanned {
			break
		}

		line := scanner.Text()
		run(line, interpreter)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(out, err)
	}
}

func RunFile(name string) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}

	interpreter := interpreter.New(os.Stdout)
	err = run(string(data), interpreter)
	if err != nil {
		return err
	}

	return nil
}

func report(err error) {
	fmt.Fprintln(os.Stderr, err)
}

func run(source string, interpreter *interpreter.Interpreter) error {
	hadError := false
	var firstError error = nil

	lexer := lexer.New(source)
	var tokens []token.Token
	for token, err := range lexer.Tokens() {
		if err != nil {
			report(err)
			hadError = true
			if firstError == nil {
				firstError = err
			}
			continue
		}

		tokens = append(tokens, token)
	}

	parser := parser.New(tokens)
	stmts, err := parser.Parse()
	if err != nil {
		report(err)
		hadError = true
		if firstError == nil {
			firstError = err
		}
	}

	if hadError {
		return firstError
	}

	err = interpreter.Interpret(stmts)
	if err != nil {
		report(err)
		hadError = true
		if firstError == nil {
			firstError = err
		}
	}

	return firstError
}
