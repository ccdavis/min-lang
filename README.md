# MinLang

A simple programming language that compiles to a stack-based virtual machine.

## Features

- Go-like syntax with functions, structs, arrays, and maps
- Infix expressions with typical operators
- Control flow: `if` and `for` statements
- Pascal-style nested functions
- Immutable (`const`) and mutable (`var`) variables
- One module per file, entry point is `main()` function

## Project Structure

```
minlang/
├── lexer/       # Lexical analysis (tokenization)
├── parser/      # Syntax analysis (AST generation)
├── ast/         # Abstract Syntax Tree definitions
├── compiler/    # Code generation (AST → bytecode)
├── vm/          # Virtual machine implementation
├── cmd/minlang/ # Main compiler executable
├── examples/    # Example programs
└── tests/       # Test files
```

## Building

```bash
go build -o minlang cmd/minlang/main.go
```

## Usage

```bash
./minlang program.min
```
