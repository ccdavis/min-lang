package main

import (
	"fmt"
	"minlang/compiler"
	"minlang/lexer"
	"minlang/parser"
	"minlang/vm"
	"os"
	"strings"
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

	bytecode := c.Bytecode()

	// Debug: print bytecode if --debug flag is present
	if len(os.Args) > 2 && os.Args[2] == "--debug" {
		fmt.Println("=== Bytecode Debug ===")
		fmt.Printf("Total constants: %d\n", len(bytecode.Constants))
		for i, constant := range bytecode.Constants {
			fmt.Printf("Constant %d: Type=%d", i, constant.Type)
			if constant.Type == 7 { // FunctionType
				fn := constant.AsFunction()
				fmt.Printf(" [Function: %s params=%d locals=%d]\n", fn.Name, fn.NumParams, fn.NumLocals)
				fmt.Println("  Function bytecode:")
				for _, line := range strings.Split(vm.Disassemble(fn.Instructions), "\n") {
					if line != "" {
						fmt.Println("   ", line)
					}
				}
			} else {
				fmt.Printf(" Value=%v\n", constant)
			}
		}
		fmt.Println("\n=== Main Bytecode ===")
		fmt.Println(vm.Disassemble(bytecode.Instructions))
		fmt.Println()
	}

	// Run
	machine := vm.New(bytecode)
	err = machine.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Runtime error: %v\n", err)
		os.Exit(1)
	}

	// Print result
	result := machine.LastPoppedStackElem()
	fmt.Println(result.String())
}
