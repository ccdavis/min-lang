package compiler

import (
	"minlang/lexer"
	"minlang/parser"
	"minlang/vm"
	"testing"
)

func TestArrayLiteral(t *testing.T) {
	input := `
var arr: []int = [1, 2, 3];
arr[0];
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

	if lastPopped.AsInt() != 1 {
		t.Fatalf("expected 1, got %d", lastPopped.AsInt())
	}
}

func TestArrayAssignment(t *testing.T) {
	input := `
var arr: []int = [1, 2, 3];
arr[1] = 10;
arr[1];
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

	if lastPopped.AsInt() != 10 {
		t.Fatalf("expected 10, got %d", lastPopped.AsInt())
	}
}

func TestArrayLen(t *testing.T) {
	input := `
var arr: []int = [1, 2, 3, 4, 5];
len(arr);
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

func TestMapLiteral(t *testing.T) {
	input := `
var m: map[string]int = map[string]int{"a": 1, "b": 2};
m["a"];
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

	if lastPopped.AsInt() != 1 {
		t.Fatalf("expected 1, got %d", lastPopped.AsInt())
	}
}

func TestMapAssignment(t *testing.T) {
	input := `
var m: map[string]int = map[string]int{"a": 1};
m["b"] = 2;
m["b"];
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

	if lastPopped.AsInt() != 2 {
		t.Fatalf("expected 2, got %d", lastPopped.AsInt())
	}
}

func TestMapLen(t *testing.T) {
	input := `
var m: map[string]int = map[string]int{"a": 1, "b": 2, "c": 3};
len(m);
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

	if lastPopped.AsInt() != 3 {
		t.Fatalf("expected 3, got %d", lastPopped.AsInt())
	}
}

func TestStructLiteral(t *testing.T) {
	input := `
var p = Person{name: "Alice", age: 30};
p.name;
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

	if lastPopped.AsString() != "Alice" {
		t.Fatalf("expected Alice, got %s", lastPopped.AsString())
	}
}

func TestStructFieldAssignment(t *testing.T) {
	input := `
var p = Person{name: "Alice", age: 30};
p.age = 31;
p.age;
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

	if lastPopped.AsInt() != 31 {
		t.Fatalf("expected 31, got %d", lastPopped.AsInt())
	}
}
