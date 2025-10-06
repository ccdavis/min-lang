package compiler

import (
	"minlang/vm"
	"testing"
)

func TestForLoops(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			// Simple while-style loop
			`var sum: int = 0;
			var i: int = 1;
			for i <= 5 {
				sum = sum + i;
				i = i + 1;
			}
			sum;`,
			15, // 1+2+3+4+5
		},
		{
			// C-style for loop
			`var sum: int = 0;
			for var i: int = 1; i <= 5; i = i + 1 {
				sum = sum + i;
			}
			sum;`,
			15,
		},
		{
			// Loop with break condition
			`var result: int = 0;
			var i: int = 0;
			for i < 100 {
				result = result + 1;
				i = i + 1;
				if i == 10 {
					i = 100;
				}
			}
			result;`,
			10,
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

func TestNestedLoops(t *testing.T) {
	input := `var sum: int = 0;
	var i: int = 1;
	for i <= 3 {
		var j: int = 1;
		for j <= 2 {
			sum = sum + 1;
			j = j + 1;
		}
		i = i + 1;
	}
	sum;`

	program := parse(input)

	compiler := New()
	err := compiler.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	vm := vm.New(compiler.Bytecode())
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	stackElem := vm.LastPoppedStackElem()
	expected := 6 // 3 iterations * 2 iterations
	testExpectedValue(t, expected, stackElem)
}
