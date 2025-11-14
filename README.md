## Intro

I created this language and the interpreters on a whim to see how far Claude Code could get me without really trying.  It's my only real "Vibe coded" project of any size. I only read the explanations of what Claude was planning to do and a few diffs it showed me, never even opening an editor. I don't actually recommend this approach.

I was pretty surprised how  competent Claude was. It helps I've made a couple of hobby or toy languages so I at least understand interpreters and compilers and know the terminology. Still I'm quite impressed with how much CC did on its own. This project would have taken me months on my own.

The whole thing took a few hours of my time. Half of that was browsing the internet while I waited.

I had Claude use Go-lang because
* It's statically typed so theAI gets more feedback to guide it as it builds
* Compile times are fast so Claude could iterate faster
* Interesting to code in a language I'm less familiar with. I know it, but I don't use it much.
* The interpreters could lean on the large Go runtime and standard library if I decided to make a standard library for Min-lang

Everything else here is all Claude Code, for better or worse.


# MinLang

A fast, educational programming language with a stack-based virtual machine. MinLang demonstrates modern compiler optimization techniques while maintaining clean, readable code.

## Features

- **Modern syntax**: Go-like syntax with type annotations
- **Rich type system**: Integers, floats, booleans, strings, arrays, maps, structs, enums
- **Functions**: First-class functions with closures and recursion
- **Control flow**: `if/else`, `for` loops, `break`, `continue`, `switch/case`
- **Variables**: Immutable (`const`) and mutable (`var`) bindings
- **Operators**: Full arithmetic, comparison, and logical operators
- **Built-in functions**: Math (`abs`, `min`, `max`, `sqrt`, `pow`, `floor`, `ceil`), String (`split`, `substring`), Collections (`len`, `append`, `keys`, `values`, `copy`, `delete`), Type conversion (`int`, `float`, `string`), and more

## Performance

MinLang achieves **~75% of Python's speed** with the register-based VM through aggressive optimizations:

- **Register-based VM**: Type-specialized opcodes, zero runtime type checks
- **Tagged union values**: Zero boxing overhead for primitives
- **Direct operations**: Peephole optimization eliminates redundant instructions
- **Frame pooling**: Zero-allocation function calls with embedded closures
- **String interning**: Memory deduplication across the program
- **Pre-allocated errors**: No allocations on error paths

The original stack-based VM achieves ~49% of Python's speed and remains available for comparison.

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
- `stdlib_demo.min` - Standard library functions showcase

## Benchmarks

Performance on heavy Mandelbrot benchmark (~82M iterations):

| Language | Time | Iterations/sec | Relative to Python |
|----------|------|----------------|---------------------|
| Python 3 | 5.3s | 15.6M/s | 1.00× (baseline) |
| MinLang (register) | 7.1s | 11.6M/s | 0.75× (75% of Python) |
| Ruby 3 | 5.8s | 14.2M/s | 0.91× (91% of Python) |
| MinLang (stack) | 10.8s | 7.6M/s | 0.49× (49% of Python) |

**Comparison summary:**
- **Register VM** achieves 75% of Python's performance - excellent for a bytecode interpreter without JIT
- Python and Ruby remain closely matched (Ruby is 91% of Python's speed)
- **Stack VM** achieves 49% of Python's performance, demonstrating the benefit of register-based architecture
- The register VM is **53% faster** than the stack VM (7.1s vs 10.8s)

All benchmarks use pure native code with no external libraries or optimizations. Use `--backend=register` (default) or `--backend=stack` to select the VM.

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
