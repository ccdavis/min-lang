# MinLang

A fast, educational programming language with a stack-based virtual machine. MinLang demonstrates modern compiler optimization techniques while maintaining clean, readable code.

## Features

- **Modern syntax**: Go-like syntax with type annotations
- **Rich type system**: Integers, floats, booleans, strings, arrays, maps, structs, enums
- **Functions**: First-class functions with closures and recursion
- **Control flow**: `if/else`, `for` loops, `break`, `continue`, `switch/case`
- **Variables**: Immutable (`const`) and mutable (`var`) bindings
- **Operators**: Full arithmetic, comparison, and logical operators
- **Built-in functions**: `print`, `len`, `push`, `pop`, `keys`

## Performance

MinLang achieves **~31% of Python's speed** through aggressive optimizations:

- Tagged union values (zero boxing overhead)
- Direct local operations (peephole optimization)
- Frame pooling and embedded closures (zero-allocation function calls)
- String interning (memory deduplication)
- Pre-allocated errors (no error path allocations)

See [PERFORMANCE.md](PERFORMANCE.md) for detailed analysis.

## Quick Start

### Build
```bash
go build -o minlang cmd/minlang/main.go
```

### Run
```bash
./minlang examples/factorial.min
```

### Debug bytecode
```bash
./minlang program.min --debug
```

## Example Program

```javascript
// Recursive fibonacci
func fib(n: int): int {
    if n <= 1 {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}

// Calculate and print
for var i: int = 0; i <= 10; i = i + 1 {
    print("fib(", i, ") =", fib(i))
}
```

## Project Structure

```
minlang/
├── lexer/       # Lexical analysis (tokenization)
├── parser/      # Syntax analysis (AST generation)
├── ast/         # Abstract Syntax Tree definitions
├── compiler/    # Code generation (bytecode + optimizations)
├── vm/          # Virtual machine (stack-based interpreter)
├── cmd/minlang/ # Main executable
├── examples/    # Example programs
└── benchmarks/  # Performance benchmarks vs Python, C, Go
```

## Architecture Highlights

### Compiler
- Single-pass compilation to bytecode
- Peephole optimization (direct local operations)
- Symbol table with scope management
- Constant folding ready

### Virtual Machine
- Stack-based architecture (2048-value stack)
- Frame pooling (pre-allocated call frames)
- Tagged union values (8-byte, no heap allocation for primitives)
- Computed dispatch with embedded closures

### Memory Management
- GC-safe object pools for heap types
- String interning for deduplication
- Pre-allocated error constants
- Zero-allocation function calls

## Language Reference

See [GRAMMAR.md](GRAMMAR.md) for complete syntax specification.

### Variable Declarations
```javascript
const x: int = 42        // Immutable
var y: float = 3.14      // Mutable
var name: string = "Bob" // Type required
```

### Functions
```javascript
func add(x: int, y: int): int {
    return x + y
}

// Closures supported
func makeCounter(): func(): int {
    var count: int = 0
    return func(): int {
        count = count + 1
        return count
    }
}
```

### Data Structures
```javascript
// Arrays
var arr: []int = [1, 2, 3, 4, 5]
print(arr[0])           // 1
print(len(arr))         // 5

// Maps
var m: map[string]int = {"a": 1, "b": 2}
print(m["a"])           // 1

// Structs
type Person struct {
    name: string
    age: int
}

var p: Person = Person{name: "Alice", age: 30}
print(p.name)           // Alice
```

### Control Flow
```javascript
// If/else
if x > 10 {
    print("big")
} else {
    print("small")
}

// For loops
for var i: int = 0; i < 10; i = i + 1 {
    print(i)
}

// While-style loops
var i: int = 0
for i < 10 {
    print(i)
    i = i + 1
}

// Switch
switch value {
case 1:
    print("one")
case 2:
    print("two")
default:
    print("other")
}
```

## Examples

The `examples/` directory contains:
- `factorial.min` - Recursive factorial
- `fibonacci.min` - Fibonacci sequence
- `mandelbrot_heavy.min` - Performance benchmark (122M iterations)
- `comprehensive_demo.min` - All language features

## Benchmarks

Performance on mandelbrot benchmark (122.3M iterations):

| Language | Time | Relative Speed |
|----------|------|----------------|
| C (gcc -O3) | 5.5s | 1.0× (baseline) |
| Go (compiled) | 8.2s | 1.5× |
| MinLang | **18.4s** | **3.3×** |
| Python 3.10 | 58.2s | 10.6× |

**MinLang is 3.16× faster than Python** on compute-intensive workloads.

## Development

### Running Tests
```bash
go test ./...
```

### Benchmark
```bash
time ./minlang examples/mandelbrot_heavy.min
```

### Performance Analysis
See [PERFORMANCE.md](PERFORMANCE.md) for:
- Detailed optimization breakdown
- GC pressure analysis
- Future optimization opportunities
- Architecture deep-dive

## Educational Value

MinLang is designed to teach:
- **Compiler construction**: Lexer → Parser → AST → Bytecode
- **VM optimization**: Tagged unions, frame pooling, peephole optimization
- **Memory management**: GC-safe unsafe.Pointer usage, object pooling
- **Performance tuning**: Profiling, bottleneck identification, incremental optimization

## License

MIT License - See LICENSE file for details.

## Credits

Built with inspiration from:
- "Writing An Interpreter In Go" by Thorsten Ball
- "Crafting Interpreters" by Robert Nystrom
- Go's runtime and compiler optimizations
