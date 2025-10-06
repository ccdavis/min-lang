package minlang_test

import (
	"bytes"
	"io"
	"minlang/compiler"
	"minlang/lexer"
	"minlang/parser"
	"minlang/vm"
	"os"
	"strings"
	"testing"
)

// runProgram is a helper that compiles and runs MinLang source code
func runProgram(t *testing.T, source string) (string, error) {
	t.Helper()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Lex
	l := lexer.New(source)

	// Parse
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		os.Stdout = oldStdout
		return "", nil
	}

	// Compile
	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		os.Stdout = oldStdout
		return "", err
	}

	// Run
	machine := vm.New(c.Bytecode())
	err = machine.Run()

	// Restore stdout and capture output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)

	if err != nil {
		return buf.String(), err
	}

	// Get result
	result := machine.LastPoppedStackElem()
	output := buf.String()

	// Append result if not nil
	if result.Type != vm.NilType {
		if output != "" {
			output += result.String() + "\n"
		} else {
			output = result.String() + "\n"
		}
	}

	return output, nil
}

// runProgramFile runs a MinLang source file
func runProgramFile(t *testing.T, filename string) (string, error) {
	t.Helper()

	source, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read %s: %v", filename, err)
	}

	return runProgram(t, string(source))
}

// TestExamplePrograms runs all example programs to ensure they don't crash
func TestExamplePrograms(t *testing.T) {
	examples := []struct {
		name           string
		file           string
		expectOutput   bool
		outputContains []string
	}{
		{"Hello", "examples/hello.min", false, []string{}},
		{"Minimal", "examples/minimal.min", false, []string{}},
		{"Arithmetic", "examples/arithmetic.min", true, []string{"230"}},
		{"Factorial", "examples/factorial.min", true, []string{"120"}},
		{"Fibonacci", "examples/fibonacci.min", true, []string{"fib(", "55"}},
		{"SimpleLoop", "examples/simple_loop.min", false, []string{}},
		{"SumLoop", "examples/sum_loop.min", false, []string{}},
		{"Conditionals", "examples/conditionals.min", false, []string{}},
		{"ConstDemo", "examples/const_demo.min", false, []string{}},
		{"ArrayDemo", "examples/array_demo.min", false, []string{}},
		{"MapDemo", "examples/map_demo.min", false, []string{}},
		{"StructDemo", "examples/struct_demo.min", false, []string{}},
		{"EnumSimple", "examples/enum_simple.min", false, []string{}},
		{"NestedFunctions", "examples/nested_functions.min", false, []string{}},
		{"SwitchSimple", "examples/switch_simple.min", false, []string{}},
		{"BreakContinueDemo", "examples/break_continue_demo.min", false, []string{}},
		{"StringOps", "examples/string_ops.min", false, []string{}},
		{"PrimeCheck", "examples/prime_check.min", false, []string{}},
		{"BuiltinsDemo", "examples/builtins_demo.min", false, []string{}},
	}

	for _, tt := range examples {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runProgramFile(t, tt.file)
			if err != nil {
				t.Fatalf("Program failed: %v", err)
			}

			if tt.expectOutput && output == "" {
				t.Error("Expected output but got none")
			}

			for _, expected := range tt.outputContains {
				if !strings.Contains(output, expected) {
					t.Errorf("Output should contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

// TestLanguageFeatures tests individual language features
func TestLanguageFeatures(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			"IntegerArithmetic",
			"print(2 + 3 * 4)",
			"14\n",
		},
		{
			"FloatArithmetic",
			"print(2.5 + 3.5)",
			"6.000000\n",
		},
		{
			"StringConcatenation",
			`print("Hello" + " " + "World")`,
			"Hello World\n",
		},
		{
			"BooleanLogic",
			"print(true && false)",
			"false\n",
		},
		{
			"Comparison",
			"print(5 > 3)",
			"true\n",
		},
		{
			"VariableAssignment",
			`var x: int = 42
print(x)`,
			"42\n",
		},
		{
			"ConstVariable",
			`const PI: float = 3.14159
print(PI)`,
			"3.141590\n",
		},
		{
			"IfStatement",
			`if 5 > 3 {
    print("yes")
} else {
    print("no")
}`,
			"yes\n",
		},
		{
			"ForLoop",
			`var sum: int = 0
for var i: int = 1; i <= 5; i = i + 1 {
    sum = sum + i
}
print(sum)`,
			"15\n",
		},
		{
			"WhileLoop",
			`var i: int = 0
var sum: int = 0
for i < 5 {
    sum = sum + i
    i = i + 1
}
print(sum)`,
			"10\n",
		},
		{
			"FunctionCall",
			`func double(x: int): int {
    return x * 2
}
print(double(21))`,
			"42\n",
		},
		{
			"RecursiveFunction",
			`func factorial(n: int): int {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}
print(factorial(5))`,
			"120\n",
		},
		{
			"ArrayLiteral",
			`var arr: []int = [1, 2, 3]
print(arr[0])
print(arr[1])
print(arr[2])`,
			"1\n2\n3\n",
		},
		{
			"ArrayLength",
			`var arr: []int = [1, 2, 3, 4, 5]
print(arr[0])
print(arr[4])`,
			"1\n5\n",
		},
		{
			"Closure",
			`func add(x: int): int {
    return x + 5
}
print(add(10))`,
			"15\n",
		},
		{
			"Break",
			`var sum: int = 0
for var i: int = 0; i < 10; i = i + 1 {
    if i == 5 {
        break
    }
    sum = sum + i
}
print(sum)`,
			"10\n",
		},
		{
			"Continue",
			`var sum: int = 0
for var i: int = 0; i < 5; i = i + 1 {
    if i == 2 {
        continue
    }
    sum = sum + i
}
print(sum)`,
			"8\n",
		},
		{
			"Negation",
			"print(-5)",
			"-5\n",
		},
		{
			"NotOperator",
			"print(!true)",
			"false\n",
		},
		{
			"IntegerEquality",
			`print(42 == 42)`,
			"true\n",
		},
		{
			"ArrayModification",
			`var arr: []int = [1, 2, 3]
arr[1] = 10
print(arr[1])`,
			"10\n",
		},
		{
			"MapStringIntBasic",
			`var ages: map[string]int = map[string]int{"Alice": 30, "Bob": 25}
print(ages["Alice"])
print(len(ages))`,
			"30\n2\n",
		},
		{
			"MapAddEntry",
			`var ages: map[string]int = map[string]int{"Alice": 30}
ages["Bob"] = 25
print(ages["Bob"])
print(len(ages))`,
			"25\n2\n",
		},
		{
			"MapModifyEntry",
			`var ages: map[string]int = map[string]int{"Alice": 30}
ages["Alice"] = 31
print(ages["Alice"])`,
			"31\n",
		},
		{
			"MapIntStringBasic",
			`var names: map[int]string = map[int]string{1: "First", 2: "Second"}
print(names[1])
print(names[2])`,
			"First\nSecond\n",
		},
		{
			"MapIntStringAdd",
			`var names: map[int]string = map[int]string{1: "First"}
names[2] = "Second"
print(names[2])
print(len(names))`,
			"Second\n2\n",
		},
		{
			"MapDelete",
			`var ages: map[string]int = map[string]int{"Alice": 30, "Bob": 25}
delete(ages, "Alice")
print(len(ages))`,
			"1\n",
		},
		{
			"ArrayStringBasic",
			`var fruits: []string = ["apple", "banana", "cherry"]
print(fruits[0])
print(fruits[1])
print(fruits[2])`,
			"apple\nbanana\ncherry\n",
		},
		{
			"ArrayFloatBasic",
			`var prices: []float = [1.5, 2.5, 3.5]
print(prices[0])
print(prices[2])`,
			"1.500000\n3.500000\n",
		},
		{
			"ArrayBoolBasic",
			`var flags: []bool = [true, false, true]
print(flags[0])
print(flags[1])
print(flags[2])`,
			"true\nfalse\ntrue\n",
		},
		{
			"ArrayNested",
			`var matrix: [][]int = [[1, 2], [3, 4]]
print(matrix[0][0])
print(matrix[0][1])
print(matrix[1][0])
print(matrix[1][1])`,
			"1\n2\n3\n4\n",
		},
		{
			"ArrayEmpty",
			`var empty: []int = []
print(len(empty))`,
			"0\n",
		},
		{
			"ArrayModifyNested",
			`var matrix: [][]int = [[1, 2], [3, 4]]
matrix[0][1] = 20
matrix[1][0] = 30
print(matrix[0][1])
print(matrix[1][0])`,
			"20\n30\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runProgram(t, tt.source)
			if err != nil {
				t.Fatalf("Program failed: %v", err)
			}

			if output != tt.expected {
				t.Errorf("Expected output:\n%s\nGot:\n%s", tt.expected, output)
			}
		})
	}
}

// TestErrorCases tests that errors are properly caught
func TestErrorCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			"DivisionByZero",
			"print(10 / 0)",
		},
		{
			"UndefinedVariable",
			"print(x)",
		},
		{
			"WrongArgumentCount",
			`func add(x: int, y: int): int { return x + y }
add(1)`,
		},
		{
			"BreakOutsideLoop",
			"break",
		},
		{
			"ContinueOutsideLoop",
			"continue",
		},
		{
			"AssignToConst",
			`const x: int = 5
x = 10`,
		},
		{
			"MapWrongKeyTypeStringForInt",
			`var ages: map[string]int = map[string]int{"Alice": 30}
var key: int = 123
print(ages[key])`,
		},
		{
			"MapWrongKeyTypeIntForString",
			`var names: map[int]string = map[int]string{1: "Alice"}
var key: string = "key"
print(names[key])`,
		},
		{
			"MapWrongValueType",
			`var ages: map[string]int = map[string]int{"Alice": 30}
ages["Bob"] = "thirty"`,
		},
		{
			"ArrayWrongElementType",
			`var nums: []int = [1, 2, 3]
nums[1] = "hello"`,
		},
		{
			"ArrayWrongLiteralType",
			`var nums: []int = [1, 2, "three"]
print(nums[0])`,
		},
		{
			"FunctionWrongArgCount",
			`func add(x: int, y: int): int { return x + y }
add(5)`,
		},
		{
			"FunctionWrongArgType",
			`func add(x: int, y: int): int { return x + y }
add(5, "hello")`,
		},
		{
			"FunctionWrongReturnType",
			`func getName(): string { return 123 }
getName()`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := runProgram(t, tt.source)
			if err == nil {
				t.Error("Expected error but got none")
			}
		})
	}
}

// TestOperatorPrecedence tests operator precedence
func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		expr     string
		expected string
	}{
		{"2 + 3 * 4", "14"},
		{"2 * 3 + 4", "10"},
		{"(2 + 3) * 4", "20"},
		{"10 - 2 - 3", "5"},
		{"2 + 3 > 4", "true"},
		{"2 * 3 == 6", "true"},
		{"true && false || true", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			source := "print(" + tt.expr + ")"
			output, err := runProgram(t, source)
			if err != nil {
				t.Fatalf("Program failed: %v", err)
			}

			expected := tt.expected + "\n"
			if output != expected {
				t.Errorf("Expected %q, got %q", expected, output)
			}
		})
	}
}

// TestBuiltinFunctions tests all built-in functions
func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected string
	}{
		{
			"Print",
			`print("hello", 42, true)`,
			"hello 42 true\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := runProgram(t, tt.source)
			if err != nil {
				t.Fatalf("Program failed: %v", err)
			}

			if output != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, output)
			}
		})
	}
}

// TestComplexPrograms tests more complex programs
func TestComplexPrograms(t *testing.T) {
	t.Run("NestedLoops", func(t *testing.T) {
		source := `var sum: int = 0
for var i: int = 1; i <= 3; i = i + 1 {
    for var j: int = 1; j <= 3; j = j + 1 {
        sum = sum + i * j
    }
}
print(sum)`

		output, err := runProgram(t, source)
		if err != nil {
			t.Fatalf("Program failed: %v", err)
		}

		// Sum: (1*1 + 1*2 + 1*3) + (2*1 + 2*2 + 2*3) + (3*1 + 3*2 + 3*3) = 6 + 12 + 18 = 36
		expected := "36\n"
		if output != expected {
			t.Errorf("Expected %q, got %q", expected, output)
		}
	})

	t.Run("NestedFunctions", func(t *testing.T) {
		source := `func outer(x: int): int {
    func inner(y: int): int {
        return x + y
    }
    return inner(10)
}
print(outer(5))`

		output, err := runProgram(t, source)
		if err != nil {
			t.Fatalf("Program failed: %v", err)
		}

		expected := "15\n"
		if output != expected {
			t.Errorf("Expected %q, got %q", expected, output)
		}
	})
}

// BenchmarkFibonacci benchmarks the fibonacci example
func BenchmarkFibonacci(b *testing.B) {
	source := `func fib(n: int): int {
    if n <= 1 {
        return n
    }
    return fib(n - 1) + fib(n - 2)
}
fib(20)`

	for i := 0; i < b.N; i++ {
		l := lexer.New(source)
		p := parser.New(l)
		program := p.ParseProgram()
		c := compiler.New()
		c.Compile(program)
		machine := vm.New(c.Bytecode())
		machine.Run()
	}
}

// BenchmarkFactorial benchmarks the factorial example
func BenchmarkFactorial(b *testing.B) {
	source := `func factorial(n: int): int {
    if n <= 1 {
        return 1
    }
    return n * factorial(n - 1)
}
factorial(10)`

	for i := 0; i < b.N; i++ {
		l := lexer.New(source)
		p := parser.New(l)
		program := p.ParseProgram()
		c := compiler.New()
		c.Compile(program)
		machine := vm.New(c.Bytecode())
		machine.Run()
	}
}
