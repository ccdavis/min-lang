package main

import (
	"flag"
	"fmt"
	"minlang/compiler"
	"minlang/lexer"
	"minlang/parser"
	"minlang/vm"
	"os"
	"runtime/pprof"
	"strings"
)

func main() {
	// Define flags
	backend := flag.String("backend", "register", "VM backend: stack or register")
	debug := flag.Bool("debug", false, "Print bytecode debug information")
	cpuprofile := flag.String("cpuprofile", "", "Write CPU profile to file")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("Usage: minlang [flags] <source-file>")
		fmt.Println("Flags:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	sourceFile := flag.Arg(0)

	// Start CPU profiling if requested
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not create CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			fmt.Fprintf(os.Stderr, "Could not start CPU profile: %v\n", err)
			os.Exit(1)
		}
		defer pprof.StopCPUProfile()
	}

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

	// Compile and run based on backend choice
	if *backend == "register" {
		// Register backend
		rc := compiler.NewRegisterCompiler()
		_, err = rc.CompileToRegister(program)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Register compilation error: %v\n", err)
			os.Exit(1)
		}

		registerBytecode := rc.RegisterBytecode()

		if *debug {
			fmt.Println("=== Register Bytecode Debug ===")
			fmt.Printf("Total constants: %d\n", len(registerBytecode.Constants))
			fmt.Printf("Max registers used: %d\n", rc.MaxRegs)
			fmt.Printf("Total instructions: %d\n", len(registerBytecode.Instructions))
			fmt.Println()
		}

		// Run register VM
		regVM := vm.NewRegisterVM(registerBytecode)
		err = regVM.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Register VM runtime error: %v\n", err)
			os.Exit(1)
		}

	} else {
		// Stack backend (default)
		c := compiler.New()
		err = c.Compile(program)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Compilation error: %v\n", err)
			os.Exit(1)
		}

		bytecode := c.Bytecode()

		// Debug: print bytecode if --debug flag is present
		if *debug {
			fmt.Println("=== Stack Bytecode Debug ===")
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

		// Run stack VM
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
}
