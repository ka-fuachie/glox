package main

import (
	"errors"
	"fmt"
	"glox/internal/cli"
	gloxerror "glox/pkg/error"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) > 1 {
		fmt.Println("Usage: glox [file]")
		os.Exit(64)
	} else if len(args) == 1 {
		err := cli.RunFile(args[0])
		if err != nil {
			if _, ok := errors.AsType[*gloxerror.StaticError](err); ok {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(65)
			} else if _, ok := errors.AsType[*gloxerror.RuntimeError](err); ok {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(70)
			} else {
				log.Fatal(err)
			}
		}
	} else {
		cli.StartRepl(os.Stdin, os.Stdout)
	}
}
