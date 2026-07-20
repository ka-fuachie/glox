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
	interpreter := interpreter.New()

	for {
		fmt.Fprint(out, "> ")

		scanned := scanner.Scan()
		if !scanned {
			break
		}

		line := scanner.Text()
		lexer := lexer.New(line)
		var tokens []token.Token
		for token, err := range lexer.Tokens() {
			if err != nil {
				report(err)
			} else {
				tokens = append(tokens, token)
			}
		}
		parser := parser.New(tokens)
		expr, err := parser.Parse()
		if err != nil {
			report(err)
			continue
		}

		result, err := interpreter.Interpret(expr)
		if err != nil {
			report(err)
			continue
		}
		fmt.Println(stringResult(result))
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

	hadError := false
	var firstError error = nil
	lexer := lexer.New(string(data))
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
	expr, err := parser.Parse()
	if err != nil {
		report(err)
		hadError = true
		if firstError == nil {
			firstError = err
		}
	}

	interpreter := interpreter.New()
	result, err := interpreter.Interpret(expr)
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

	fmt.Println(stringResult(result))

	return nil
}

func report(err error) {
	fmt.Fprintln(os.Stderr, err)
}

func stringResult(result any) string {
	if result == nil {
		return "nil"
	}

	return fmt.Sprintf("%#v", result)
}
