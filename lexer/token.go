package lexer

import "fmt"

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	EOF TokenType = iota
	ILLEGAL

	// Identifiers and literals
	IDENT  // foo, bar, x, y
	INT    // 123
	FLOAT  // 123.456
	STRING // "hello"

	// Keywords
	VAR
	CONST
	FUNC
	TYPE
	STRUCT
	ENUM
	RETURN
	IF
	ELSE
	FOR
	BREAK
	CONTINUE
	SWITCH
	CASE
	DEFAULT
	MAP
	TRUE
	FALSE
	NIL

	// Operators
	PLUS     // +
	MINUS    // -
	ASTERISK // *
	SLASH    // /
	PERCENT  // %

	EQ     // ==
	NE     // !=
	LT     // <
	GT     // >
	LE     // <=
	GE     // >=
	AND    // &&
	OR     // ||
	NOT    // !
	ASSIGN // =

	// Delimiters
	COLON     // :
	SEMICOLON // ;
	COMMA     // ,
	DOT       // .

	LPAREN   // (
	RPAREN   // )
	LBRACE   // {
	RBRACE   // }
	LBRACKET // [
	RBRACKET // ]
)

var keywords = map[string]TokenType{
	"var":      VAR,
	"const":    CONST,
	"func":     FUNC,
	"type":     TYPE,
	"struct":   STRUCT,
	"enum":     ENUM,
	"return":   RETURN,
	"if":       IF,
	"else":     ELSE,
	"for":      FOR,
	"break":    BREAK,
	"continue": CONTINUE,
	"switch":   SWITCH,
	"case":     CASE,
	"default":  DEFAULT,
	"map":      MAP,
	"true":     TRUE,
	"false":    FALSE,
	"nil":      NIL,
}

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// String returns a string representation of the token type
func (t TokenType) String() string {
	switch t {
	case EOF:
		return "EOF"
	case ILLEGAL:
		return "ILLEGAL"
	case IDENT:
		return "IDENT"
	case INT:
		return "INT"
	case FLOAT:
		return "FLOAT"
	case STRING:
		return "STRING"
	case VAR:
		return "VAR"
	case CONST:
		return "CONST"
	case FUNC:
		return "FUNC"
	case TYPE:
		return "TYPE"
	case STRUCT:
		return "STRUCT"
	case RETURN:
		return "RETURN"
	case IF:
		return "IF"
	case ELSE:
		return "ELSE"
	case FOR:
		return "FOR"
	case BREAK:
		return "BREAK"
	case CONTINUE:
		return "CONTINUE"
	case SWITCH:
		return "SWITCH"
	case CASE:
		return "CASE"
	case DEFAULT:
		return "DEFAULT"
	case ENUM:
		return "ENUM"
	case MAP:
		return "MAP"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	case NIL:
		return "NIL"
	case PLUS:
		return "+"
	case MINUS:
		return "-"
	case ASTERISK:
		return "*"
	case SLASH:
		return "/"
	case PERCENT:
		return "%"
	case EQ:
		return "=="
	case NE:
		return "!="
	case LT:
		return "<"
	case GT:
		return ">"
	case LE:
		return "<="
	case GE:
		return ">="
	case AND:
		return "&&"
	case OR:
		return "||"
	case NOT:
		return "!"
	case ASSIGN:
		return "="
	case COLON:
		return ":"
	case SEMICOLON:
		return ";"
	case COMMA:
		return ","
	case DOT:
		return "."
	case LPAREN:
		return "("
	case RPAREN:
		return ")"
	case LBRACE:
		return "{"
	case RBRACE:
		return "}"
	case LBRACKET:
		return "["
	case RBRACKET:
		return "]"
	default:
		return "UNKNOWN"
	}
}

// String returns a string representation of the token
func (t Token) String() string {
	return fmt.Sprintf("%s '%s' (Line %d, Col %d)", t.Type.String(), t.Literal, t.Line, t.Column)
}

// LookupIdent checks if an identifier is a keyword
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
