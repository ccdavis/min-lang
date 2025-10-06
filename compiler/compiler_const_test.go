package compiler

import (
	"minlang/lexer"
	"minlang/parser"
	"strings"
	"testing"
)

func TestConstEnforcement(t *testing.T) {
	input := `
const x: int = 5;
x = 10;
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
		t.Fatalf("expected compilation error for const reassignment, got none")
	}

	if !strings.Contains(err.Error(), "cannot assign to const variable") {
		t.Fatalf("expected error about const variable, got: %s", err.Error())
	}
}

func TestVarMutable(t *testing.T) {
	input := `
var x: int = 5;
x = 10;
x;
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
		t.Fatalf("unexpected compilation error: %s", err)
	}
}

func TestConstInLoop(t *testing.T) {
	input := `
const max: int = 10;
for var i: int = 0; i < max; i = i + 1 {
	var x: int = i;
}
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
		t.Fatalf("unexpected compilation error: %s", err)
	}
}

func TestConstReassignInFunction(t *testing.T) {
	input := `
func test(): int {
	const x: int = 5;
	x = 10;
	return x;
}
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
		t.Fatalf("expected compilation error for const reassignment in function, got none")
	}

	if !strings.Contains(err.Error(), "cannot assign to const variable") {
		t.Fatalf("expected error about const variable, got: %s", err.Error())
	}
}
