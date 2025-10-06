package lexer

import (
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `var x: int = 5;
const y = 10;

func add(a: int, b: int): int {
	return a + b;
}

struct Point {
	x: int;
	y: int;
}

// This is a comment
/* This is a
   block comment */

if x == 5 {
	x = x + 1;
}

for i := 0; i < 10; i = i + 1 {
	// loop body
}

var arr: []int = [1, 2, 3];
var m: map[string]int = map[string]int{"a": 1};
var p: Point = Point{x: 1, y: 2};

x != y
x < y
x > y
x <= y
x >= y
x && y
x || y
!x
3.14
"hello world"
true
false
nil
`

	tests := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{VAR, "var"},
		{IDENT, "x"},
		{COLON, ":"},
		{IDENT, "int"},
		{ASSIGN, "="},
		{INT, "5"},
		{SEMICOLON, ";"},
		{CONST, "const"},
		{IDENT, "y"},
		{ASSIGN, "="},
		{INT, "10"},
		{SEMICOLON, ";"},
		{FUNC, "func"},
		{IDENT, "add"},
		{LPAREN, "("},
		{IDENT, "a"},
		{COLON, ":"},
		{IDENT, "int"},
		{COMMA, ","},
		{IDENT, "b"},
		{COLON, ":"},
		{IDENT, "int"},
		{RPAREN, ")"},
		{COLON, ":"},
		{IDENT, "int"},
		{LBRACE, "{"},
		{RETURN, "return"},
		{IDENT, "a"},
		{PLUS, "+"},
		{IDENT, "b"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},
		{STRUCT, "struct"},
		{IDENT, "Point"},
		{LBRACE, "{"},
		{IDENT, "x"},
		{COLON, ":"},
		{IDENT, "int"},
		{SEMICOLON, ";"},
		{IDENT, "y"},
		{COLON, ":"},
		{IDENT, "int"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},
		{IF, "if"},
		{IDENT, "x"},
		{EQ, "=="},
		{INT, "5"},
		{LBRACE, "{"},
		{IDENT, "x"},
		{ASSIGN, "="},
		{IDENT, "x"},
		{PLUS, "+"},
		{INT, "1"},
		{SEMICOLON, ";"},
		{RBRACE, "}"},
		{FOR, "for"},
		{IDENT, "i"},
		{COLON, ":"},
		{ASSIGN, "="},
		{INT, "0"},
		{SEMICOLON, ";"},
		{IDENT, "i"},
		{LT, "<"},
		{INT, "10"},
		{SEMICOLON, ";"},
		{IDENT, "i"},
		{ASSIGN, "="},
		{IDENT, "i"},
		{PLUS, "+"},
		{INT, "1"},
		{LBRACE, "{"},
		{RBRACE, "}"},
		{VAR, "var"},
		{IDENT, "arr"},
		{COLON, ":"},
		{LBRACKET, "["},
		{RBRACKET, "]"},
		{IDENT, "int"},
		{ASSIGN, "="},
		{LBRACKET, "["},
		{INT, "1"},
		{COMMA, ","},
		{INT, "2"},
		{COMMA, ","},
		{INT, "3"},
		{RBRACKET, "]"},
		{SEMICOLON, ";"},
		{VAR, "var"},
		{IDENT, "m"},
		{COLON, ":"},
		{MAP, "map"},
		{LBRACKET, "["},
		{IDENT, "string"},
		{RBRACKET, "]"},
		{IDENT, "int"},
		{ASSIGN, "="},
		{MAP, "map"},
		{LBRACKET, "["},
		{IDENT, "string"},
		{RBRACKET, "]"},
		{IDENT, "int"},
		{LBRACE, "{"},
		{STRING, "a"},
		{COLON, ":"},
		{INT, "1"},
		{RBRACE, "}"},
		{SEMICOLON, ";"},
		{VAR, "var"},
		{IDENT, "p"},
		{COLON, ":"},
		{IDENT, "Point"},
		{ASSIGN, "="},
		{IDENT, "Point"},
		{LBRACE, "{"},
		{IDENT, "x"},
		{COLON, ":"},
		{INT, "1"},
		{COMMA, ","},
		{IDENT, "y"},
		{COLON, ":"},
		{INT, "2"},
		{RBRACE, "}"},
		{SEMICOLON, ";"},
		{IDENT, "x"},
		{NE, "!="},
		{IDENT, "y"},
		{IDENT, "x"},
		{LT, "<"},
		{IDENT, "y"},
		{IDENT, "x"},
		{GT, ">"},
		{IDENT, "y"},
		{IDENT, "x"},
		{LE, "<="},
		{IDENT, "y"},
		{IDENT, "x"},
		{GE, ">="},
		{IDENT, "y"},
		{IDENT, "x"},
		{AND, "&&"},
		{IDENT, "y"},
		{IDENT, "x"},
		{OR, "||"},
		{IDENT, "y"},
		{NOT, "!"},
		{IDENT, "x"},
		{FLOAT, "3.14"},
		{STRING, "hello world"},
		{TRUE, "true"},
		{FALSE, "false"},
		{NIL, "nil"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLineAndColumn(t *testing.T) {
	input := `var x = 5;
var y = 10;`

	tests := []struct {
		expectedLine   int
		expectedColumn int
	}{
		{1, 1}, // var
		{1, 5}, // x
		{1, 7}, // =
		{1, 9}, // 5
		{1, 10}, // ;
		{2, 1}, // var
		{2, 5}, // y
		{2, 7}, // =
		{2, 9}, // 10
		{2, 11}, // ;
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Line != tt.expectedLine {
			t.Fatalf("tests[%d] - line wrong. expected=%d, got=%d",
				i, tt.expectedLine, tok.Line)
		}

		if tok.Column != tt.expectedColumn {
			t.Fatalf("tests[%d] - column wrong. expected=%d, got=%d",
				i, tt.expectedColumn, tok.Column)
		}
	}
}
