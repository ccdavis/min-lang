package compiler

import (
	"minlang/ast"
	"minlang/lexer"
	"minlang/parser"
	"minlang/vm"
	"testing"
)

func TestIntegerArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"2 * 3", 6},
		{"6 / 2", 3},
		{"10 % 3", 1},
		{"-5", -5},
		{"5 + 5 + 5", 15},
		{"2 * 2 * 2", 8},
		{"5 + 2 * 3", 11},
		{"(5 + 2) * 3", 21},
	}

	for _, tt := range tests {
		program := parse(tt.input)

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

		testExpectedValue(t, tt.expected, stackElem)
	}
}

func TestBooleanExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!!true", true},
		{"!!false", false},
	}

	for _, tt := range tests {
		program := parse(tt.input)

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

		testExpectedValue(t, tt.expected, stackElem)
	}
}

func TestConditionals(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"var x: int = 1; if true { x = 10; } x;", 10},
		{"var x: int = 0; if true { x = 10; } else { x = 20; } x;", 10},
		{"var x: int = 0; if false { x = 10; } else { x = 20; } x;", 20},
		{"var x: int = 0; if 1 < 2 { x = 10; } x;", 10},
		{"var x: int = 0; if 1 < 2 { x = 10; } else { x = 20; } x;", 10},
		{"var x: int = 0; if 1 > 2 { x = 10; } else { x = 20; } x;", 20},
	}

	for _, tt := range tests {
		program := parse(tt.input)

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

		testExpectedValue(t, tt.expected, stackElem)
	}
}

func TestGlobalVarStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"var one: int = 1; one;", 1},
		{"var one: int = 1; var two: int = 2; one + two;", 3},
		{"var one: int = 1; var two: int = one + one; one + two;", 3},
	}

	for _, tt := range tests {
		program := parse(tt.input)

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

		testExpectedValue(t, tt.expected, stackElem)
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func testExpectedValue(t *testing.T, expected interface{}, actual vm.Value) {
	t.Helper()

	switch expected := expected.(type) {
	case int:
		testIntegerValue(t, int64(expected), actual)
	case bool:
		testBooleanValue(t, expected, actual)
	}
}

func testIntegerValue(t *testing.T, expected int64, actual vm.Value) {
	t.Helper()

	if actual.Type != vm.IntType {
		t.Errorf("value type is not IntType. got=%d", actual.Type)
		return
	}

	if actual.AsInt() != expected {
		t.Errorf("value has wrong value. got=%d, want=%d", actual.AsInt(), expected)
	}
}

func testBooleanValue(t *testing.T, expected bool, actual vm.Value) {
	t.Helper()

	if actual.Type != vm.BoolType {
		t.Errorf("value type is not BoolType. got=%d", actual.Type)
		return
	}

	if actual.AsBool() != expected {
		t.Errorf("value has wrong value. got=%t, want=%t", actual.AsBool(), expected)
	}
}
