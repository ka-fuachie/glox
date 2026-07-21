package cli

import (
	"bufio"
	"fmt"
	gloxError "glox/pkg/error"
	"glox/pkg/interpreter"
	"glox/pkg/lexer"
	"glox/pkg/parser"
	"io"
	"os"
	"slices"
)

func StartRepl(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	errorReporter := gloxError.NewReporter(func(_ error) {})
	interpreter := interpreter.New(out, errorReporter)

	for {
		fmt.Fprint(out, "> ")

		scanned := scanner.Scan()
		if !scanned {
			break
		}

		line := scanner.Text()
		run(line, interpreter, errorReporter)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(out, err)
	}
}

func RunFile(name string) (bool, bool, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return false, false, err
	}

	hadError := false
	hadRuntimeError := false
	staticErrorReporter := gloxError.NewReporter(func(_ error) {
		hadError = true
	})
	runtimeErrorReporter := gloxError.NewReporter(func(_ error) {
		hadRuntimeError = true
	})

	interpreter := interpreter.New(os.Stdout, runtimeErrorReporter)
	run(string(data), interpreter, staticErrorReporter)
	if hadError || hadRuntimeError {
		return hadError, hadRuntimeError, nil
	}

	return false, false, nil
}

func run(
	source string,
	interpreter *interpreter.Interpreter,
	staticErrorReporter *gloxError.Reporter,
) (bool, bool) {
	hadError := false
	hadRuntimeError := false

	lexer := lexer.New(source, staticErrorReporter)
	tokens := slices.Collect(lexer.Tokens())
	parser := parser.New(tokens, staticErrorReporter)
	stmts := parser.Parse()

	if hadError {
		return hadError, hadRuntimeError
	}

	interpreter.Interpret(stmts)

	return hadError, hadRuntimeError
}
