package main

import (
	"fmt"
	"minlang/compiler"
	"minlang/lexer"
	"minlang/parser"
	"minlang/vm"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: minlang <source-file>")
		os.Exit(1)
	}

	sourceFile := os.Args[1]

	// Read source file
	source, err := os.ReadFile(sourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Lex
	l := lexer.New(string(source))

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		fmt.Fprintln(os.Stderr, "Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "\t%s\n", msg)
		}
		os.Exit(1)
	}

	// Compile
	c := compiler.New()
	err = c.Compile(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation error: %v\n", err)
		os.Exit(1)
	}

	// Run
	machine := vm.New(c.Bytecode())
	err = machine.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}

	// Print result
	result := machine.LastPoppedStackElem()
	fmt.Println(result.String())
}
