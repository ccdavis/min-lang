package compiler

import (
	"minlang/lexer"
	"minlang/parser"
	"minlang/vm"
	"testing"
)

func TestAppendBuiltin(t *testing.T) {
	input := `
var arr: []int = [1, 2, 3];
var arr2 = append(arr, 4);
len(arr2);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	machine := vm.New(c.Bytecode())
	err = machine.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	lastPopped := machine.LastPoppedStackElem()
	if lastPopped.Type != vm.IntType {
		t.Fatalf("expected int type, got %d", lastPopped.Type)
	}

	if lastPopped.AsInt() != 4 {
		t.Fatalf("expected 4, got %d", lastPopped.AsInt())
	}
}

func TestAppendMultipleValues(t *testing.T) {
	input := `
var arr: []int = [1, 2];
var arr2 = append(arr, 3, 4, 5);
len(arr2);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	machine := vm.New(c.Bytecode())
	err = machine.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	lastPopped := machine.LastPoppedStackElem()
	if lastPopped.Type != vm.IntType {
		t.Fatalf("expected int type, got %d", lastPopped.Type)
	}

	if lastPopped.AsInt() != 5 {
		t.Fatalf("expected 5, got %d", lastPopped.AsInt())
	}
}

func TestStringIndexing(t *testing.T) {
	input := `
var str: string = "hello";
str[0];
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	machine := vm.New(c.Bytecode())
	err = machine.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	lastPopped := machine.LastPoppedStackElem()
	if lastPopped.Type != vm.StringType {
		t.Fatalf("expected string type, got %d", lastPopped.Type)
	}

	if lastPopped.AsString() != "h" {
		t.Fatalf("expected 'h', got %s", lastPopped.AsString())
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `
var str1: string = "Hello, ";
var str2: string = "World!";
var result = str1 + str2;
result;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	machine := vm.New(c.Bytecode())
	err = machine.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	lastPopped := machine.LastPoppedStackElem()
	if lastPopped.Type != vm.StringType {
		t.Fatalf("expected string type, got %d", lastPopped.Type)
	}

	if lastPopped.AsString() != "Hello, World!" {
		t.Fatalf("expected 'Hello, World!', got %s", lastPopped.AsString())
	}
}

func TestStringAndNumberConcatenation(t *testing.T) {
	input := `
var str: string = "Count: ";
var num: int = 42;
var result = str + num;
result;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	machine := vm.New(c.Bytecode())
	err = machine.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	lastPopped := machine.LastPoppedStackElem()
	if lastPopped.Type != vm.StringType {
		t.Fatalf("expected string type, got %d", lastPopped.Type)
	}

	if lastPopped.AsString() != "Count: 42" {
		t.Fatalf("expected 'Count: 42', got %s", lastPopped.AsString())
	}
}

func TestStringLen(t *testing.T) {
	input := `
var str: string = "hello";
len(str);
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}

	machine := vm.New(c.Bytecode())
	err = machine.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}

	lastPopped := machine.LastPoppedStackElem()
	if lastPopped.Type != vm.IntType {
		t.Fatalf("expected int type, got %d", lastPopped.Type)
	}

	if lastPopped.AsInt() != 5 {
		t.Fatalf("expected 5, got %d", lastPopped.AsInt())
	}
}
