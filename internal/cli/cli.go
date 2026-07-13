package cli

import (
	"bufio"
	"errors"
	"fmt"
	"glox/pkg/lexer"
	"io"
	"os"
)

func StartRepl(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprint(out, "> ")

		scanned := scanner.Scan()
		if !scanned {
			break
		}

		line := scanner.Text()
		lexer := lexer.New(line)
		for token, err := range lexer.Tokens() {
			if err != nil {
				report(err)
			} else {
				fmt.Fprintln(out, token)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(out, err)
	}
}

var EvalError = errors.New("Syntax error")

func RunFile(name string) error {
	data, err := os.ReadFile(name)
	if err != nil {
		return err
	}

	hadError := false
	var firstError error = nil
	lexer := lexer.New(string(data))
	for token, err := range lexer.Tokens() {
		if err != nil {
			report(err)
			hadError = true
			continue
		}

		fmt.Println(token)
	}

	if hadError {
		return firstError
	}
	return nil
}

func report(err error) {
	fmt.Fprintln(os.Stderr, err)
}
