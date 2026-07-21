package main

import (
	"fmt"
	"glox/internal/cli"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) > 1 {
		fmt.Println("Usage: glox [file]")
		os.Exit(64)
	} else if len(args) == 1 {
		hadError, hadRuntimeError, err := cli.RunFile(args[0])
		if err != nil {
			log.Fatal(err)
		}

		if hadError {
			os.Exit(65)
		} else if hadRuntimeError {
			os.Exit(70)
		}
	} else {
		cli.StartRepl(os.Stdin, os.Stdout)
	}
}
