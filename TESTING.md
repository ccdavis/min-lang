# MinLang Test Suite

Comprehensive test suite for MinLang using standard Go testing practices.

## Running Tests

### Quick Test
```bash
go test ./... -v
```

### With Coverage
```bash
go test ./lexer ./parser ./compiler ./vm . -cover
```

### Using Test Runner Script
```bash
./test.sh
```

## Test Structure

### Integration Tests (`integration_test.go`)
Comprehensive end-to-end tests that exercise the entire language pipeline:

**Test Categories:**
- `TestExamplePrograms` - Runs all 19 example programs to ensure no regressions
- `TestLanguageFeatures` - Tests 18 core language features individually
- `TestOperatorPrecedence` - Verifies correct operator precedence
- `TestBuiltinFunctions` - Tests built-in functions
- `TestErrorCases` - Ensures errors are properly caught
- `TestComplexPrograms` - Tests nested constructs and complex scenarios

**Coverage:**
- ✅ Arithmetic operations (int, float)
- ✅ String concatenation
- ✅ Boolean logic
- ✅ Comparisons
- ✅ Variables (var, const)
- ✅ Control flow (if/else, for loops)
- ✅ Functions (regular, recursive)
- ✅ Arrays (literals, indexing, modification)
- ✅ Break and continue statements
- ✅ Negation and logical NOT
- ✅ Closures and nested functions

### Unit Tests

#### Lexer Tests (`lexer/lexer_test.go`)
- Token identification
- String and number literals
- Keywords and operators
- Coverage: 65.0%

#### Parser Tests (`parser/parser_test.go`)
- AST generation
- Expression parsing
- Statement parsing
- Coverage: 39.1%

#### Compiler Tests (`compiler/*_test.go`)
- Bytecode generation
- Symbol table management
- Scope handling
- Const/loop/function/break-continue compilation
- Coverage: 52.3%

#### VM Tests (`vm/vm_test.go`, `vm/vm_comprehensive_test.go`)
- Value type operations
- String interning
- Array/Map/Struct operations
- GC protection pools
- Stack operations
- Pre-allocated errors
- Coverage: 21.6%

## Test Features

### Example Programs Tested
All 19 example programs are regression tested:
- arithmetic.min
- factorial.min (recursive)
- fibonacci.min (recursive)
- array_demo.min
- map_demo.min
- struct_demo.min
- enum_simple.min
- nested_functions.min
- switch_simple.min
- break_continue_demo.min
- string_ops.min
- prime_check.min
- builtins_demo.min
- And more...

### Error Testing
Verifies proper error handling for:
- Division by zero
- Undefined variables
- Wrong argument counts
- Break/continue outside loops
- Assignment to const

### Benchmarks
Performance benchmarks included:
- `BenchmarkFibonacci` - Recursive fibonacci
- `BenchmarkFactorial` - Recursive factorial

Run benchmarks with:
```bash
go test . -bench=. -benchtime=5s
```

## Writing New Tests

### Integration Test Template
```go
{
    "TestName",
    `source code here`,
    "expected output\n",
},
```

### Adding Example Program Test
```go
{"Name", "examples/file.min", false, []string{}},
```

## Continuous Integration

The test suite is designed for CI/CD integration:
- Fast execution (< 1 second for full suite)
- Clear pass/fail indicators
- Coverage tracking
- No external dependencies

## Test Coverage Goals

| Component | Current | Target |
|-----------|---------|--------|
| Lexer | 65.0% | 80%+ |
| Parser | 39.1% | 70%+ |
| Compiler | 52.3% | 75%+ |
| VM | 21.6% | 60%+ |
| Integration | High | Maintain |

## Known Test Limitations

Some language features are tested via integration tests but may not have dedicated unit tests:
- Map operations with string keys
- Complex closure scenarios
- Switch statement edge cases
- Some built-in functions (len, push, pop not yet implemented)

These represent opportunities for future enhancement rather than gaps in correctness verification.

## Contributing Tests

When adding new language features:
1. Add integration test to `integration_test.go`
2. Add example program to `examples/`
3. Add unit tests to appropriate `*_test.go` file
4. Run full test suite before committing
5. Ensure coverage doesn't decrease

## Test Output Example

```
=== RUN   TestLanguageFeatures
=== RUN   TestLanguageFeatures/IntegerArithmetic
=== RUN   TestLanguageFeatures/FloatArithmetic
=== RUN   TestLanguageFeatures/FunctionCall
=== RUN   TestLanguageFeatures/RecursiveFunction
--- PASS: TestLanguageFeatures (0.00s)
    --- PASS: TestLanguageFeatures/IntegerArithmetic (0.00s)
    --- PASS: TestLanguageFeatures/FloatArithmetic (0.00s)
    --- PASS: TestLanguageFeatures/FunctionCall (0.00s)
    --- PASS: TestLanguageFeatures/RecursiveFunction (0.00s)
PASS
ok      minlang    0.010s
```

## Debugging Failed Tests

### Verbose Output
```bash
go test ./... -v
```

### Run Specific Test
```bash
go test . -run TestLanguageFeatures/RecursiveFunction -v
```

### Show Test Coverage
```bash
go test . -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Debug Integration Test
Add debug flag to see bytecode:
```go
output, err := runProgram(t, source+" --debug")
```

## Test Maintenance

- Tests run on every commit
- Example programs verified working
- Regression prevention built-in
- Coverage tracked over time
