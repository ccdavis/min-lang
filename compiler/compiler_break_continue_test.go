package compiler

import (
	"minlang/lexer"
	"minlang/parser"
	"minlang/vm"
	"testing"
)

func TestBreakStatement(t *testing.T) {
	input := `
var sum: int = 0;
for var i: int = 0; i < 10; i = i + 1 {
	sum = sum + i;
	if i == 5 {
		break;
	}
}
sum;
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

	// sum should be 0+1+2+3+4+5 = 15
	if lastPopped.AsInt() != 15 {
		t.Fatalf("expected 15, got %d", lastPopped.AsInt())
	}
}

func TestContinueStatement(t *testing.T) {
	input := `
var sum: int = 0;
for var i: int = 0; i < 10; i = i + 1 {
	if i == 5 {
		continue;
	}
	sum = sum + i;
}
sum;
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

	// sum should be 0+1+2+3+4+6+7+8+9 = 40 (skipped 5)
	if lastPopped.AsInt() != 40 {
		t.Fatalf("expected 40, got %d", lastPopped.AsInt())
	}
}

func TestBreakAndContinueTogether(t *testing.T) {
	input := `
var sum: int = 0;
for var i: int = 0; i < 100; i = i + 1 {
	if i == 20 {
		break;
	}
	if i % 2 == 0 {
		continue;
	}
	sum = sum + i;
}
sum;
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

	// sum should be 1+3+5+7+9+11+13+15+17+19 = 100
	if lastPopped.AsInt() != 100 {
		t.Fatalf("expected 100, got %d", lastPopped.AsInt())
	}
}

func TestBreakOutsideLoop(t *testing.T) {
	input := `
var x: int = 5;
break;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	err := c.Compile(program)
	if err == nil {
		t.Fatalf("expected compilation error for break outside loop, got none")
	}
}

func TestContinueOutsideLoop(t *testing.T) {
	input := `
var x: int = 5;
continue;
`

	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	c := New()
	err := c.Compile(program)
	if err == nil {
		t.Fatalf("expected compilation error for continue outside loop, got none")
	}
}
