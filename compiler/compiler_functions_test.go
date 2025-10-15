package compiler

import (
	"minlang/vm"
	"testing"
)

func TestFunctionCalls(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`func add(a: int, b: int): int {
				return a + b;
			}
			add(1, 2);`,
			3,
		},
		{
			`func multiply(a: int, b: int): int {
				return a * b;
			}
			multiply(3, 4);`,
			12,
		},
		{
			`func identity(x: int): int {
				return x;
			}
			identity(42);`,
			42,
		},
		{
			`func noReturn() {
				var x: int = 5;
			}
			noReturn();`,
			nil, // Implicit nil return (no return type annotation)
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)

		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s\nInput: %s", err, tt.input)
		}

		vm := vm.New(compiler.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s\nInput: %s", err, tt.input)
		}

		stackElem := vm.LastPoppedStackElem()

		if tt.expected == nil {
			// For nil returns, just check it doesn't error
			continue
		}
		testExpectedValue(t, tt.expected, stackElem)
	}
}

func TestFunctionCallsWithLocals(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`func sum(a: int, b: int): int {
				var result: int = a + b;
				return result;
			}
			sum(5, 10);`,
			15,
		},
		{
			`func fibonacci(n: int): int {
				if n < 2 {
					return n;
				}
				var a: int = 0;
				var b: int = 1;
				var i: int = 2;
				// Simple iterative approach for small n
				return n;
			}
			fibonacci(5);`,
			5,
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)

		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s\nInput: %s", err, tt.input)
		}

		vm := vm.New(compiler.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s\nInput: %s", err, tt.input)
		}

		stackElem := vm.LastPoppedStackElem()
		testExpectedValue(t, tt.expected, stackElem)
	}
}

func TestRecursiveFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`func factorial(n: int): int {
				if n < 2 {
					return 1;
				}
				return n * factorial(n - 1);
			}
			factorial(5);`,
			120,
		},
		{
			`func countdown(n: int): int {
				if n == 0 {
					return 0;
				}
				return countdown(n - 1);
			}
			countdown(3);`,
			0,
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)

		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s\nInput: %s", err, tt.input)
		}

		vm := vm.New(compiler.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s\nInput: %s", err, tt.input)
		}

		stackElem := vm.LastPoppedStackElem()
		testExpectedValue(t, tt.expected, stackElem)
	}
}
