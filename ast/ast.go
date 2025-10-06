package ast

import (
	"minlang/lexer"
	"strings"
)

// Node is the base interface for all AST nodes
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a statement node
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of the AST
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out strings.Builder
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// Identifier represents an identifier
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral represents an integer literal
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral represents a float literal
type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral represents a string literal
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

// BooleanLiteral represents a boolean literal
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// NilLiteral represents a nil literal
type NilLiteral struct {
	Token lexer.Token
}

func (nl *NilLiteral) expressionNode()      {}
func (nl *NilLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NilLiteral) String() string       { return "nil" }

// PrefixExpression represents a prefix expression (e.g., -x, !x)
type PrefixExpression struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

// InfixExpression represents an infix expression (e.g., x + y)
type InfixExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

// CallExpression represents a function call
type CallExpression struct {
	Token     lexer.Token // The '(' token
	Function  Expression  // Identifier or function expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var args []string
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	return ce.Function.String() + "(" + strings.Join(args, ", ") + ")"
}

// IndexExpression represents array/map indexing
type IndexExpression struct {
	Token lexer.Token // The '[' token
	Left  Expression  // The array or map
	Index Expression  // The index
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	return "(" + ie.Left.String() + "[" + ie.Index.String() + "])"
}

// FieldAccessExpression represents field access (e.g., x.y)
type FieldAccessExpression struct {
	Token lexer.Token // The '.' token
	Left  Expression  // The object
	Field *Identifier // The field name
}

func (fae *FieldAccessExpression) expressionNode()      {}
func (fae *FieldAccessExpression) TokenLiteral() string { return fae.Token.Literal }
func (fae *FieldAccessExpression) String() string {
	return "(" + fae.Left.String() + "." + fae.Field.String() + ")"
}

// ArrayLiteral represents an array literal
type ArrayLiteral struct {
	Token    lexer.Token // The '[' token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var elements []string
	for _, e := range al.Elements {
		elements = append(elements, e.String())
	}
	return "[" + strings.Join(elements, ", ") + "]"
}

// MapLiteral represents a map literal
type MapLiteral struct {
	Token lexer.Token // The 'map' token
	KeyType   *TypeAnnotation
	ValueType *TypeAnnotation
	Pairs     map[Expression]Expression
}

func (ml *MapLiteral) expressionNode()      {}
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Literal }
func (ml *MapLiteral) String() string {
	var pairs []string
	for k, v := range ml.Pairs {
		pairs = append(pairs, k.String()+": "+v.String())
	}
	return "map[" + ml.KeyType.String() + "]" + ml.ValueType.String() + "{" + strings.Join(pairs, ", ") + "}"
}

// StructLiteral represents a struct literal
type StructLiteral struct {
	Token  lexer.Token // The struct name token
	Name   *Identifier
	Fields map[string]Expression
}

func (sl *StructLiteral) expressionNode()      {}
func (sl *StructLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StructLiteral) String() string {
	var fields []string
	for k, v := range sl.Fields {
		fields = append(fields, k+": "+v.String())
	}
	return sl.Name.String() + "{" + strings.Join(fields, ", ") + "}"
}

// TypeAnnotation represents a type annotation
type TypeAnnotation struct {
	Token lexer.Token
	Name  string // e.g., "int", "float", "string", "bool"
	// For complex types
	IsArray     bool
	IsMap       bool
	IsFunction  bool
	ElementType *TypeAnnotation   // For arrays
	KeyType     *TypeAnnotation   // For maps
	ValueType   *TypeAnnotation   // For maps and function returns
	ParamTypes  []*TypeAnnotation // For functions
}

func (ta *TypeAnnotation) String() string {
	if ta.IsArray {
		return "[]" + ta.ElementType.String()
	}
	if ta.IsMap {
		return "map[" + ta.KeyType.String() + "]" + ta.ValueType.String()
	}
	if ta.IsFunction {
		var params []string
		for _, p := range ta.ParamTypes {
			params = append(params, p.String())
		}
		ret := ""
		if ta.ValueType != nil {
			ret = ta.ValueType.String()
		}
		return "func(" + strings.Join(params, ", ") + ") " + ret
	}
	return ta.Name
}

// VarStatement represents a variable declaration
type VarStatement struct {
	Token      lexer.Token // The 'var' token
	Name       *Identifier
	Type       *TypeAnnotation
	Value      Expression
	IsMutable  bool
}

func (vs *VarStatement) statementNode()       {}
func (vs *VarStatement) TokenLiteral() string { return vs.Token.Literal }
func (vs *VarStatement) String() string {
	keyword := "var"
	if !vs.IsMutable {
		keyword = "const"
	}
	out := keyword + " " + vs.Name.String()
	if vs.Type != nil {
		out += ": " + vs.Type.String()
	}
	if vs.Value != nil {
		out += " = " + vs.Value.String()
	}
	return out + ";"
}

// AssignmentStatement represents an assignment
type AssignmentStatement struct {
	Token lexer.Token // The '=' token
	Left  Expression  // Can be Identifier, IndexExpression, or FieldAccessExpression
	Value Expression
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignmentStatement) String() string {
	return as.Left.String() + " = " + as.Value.String() + ";"
}

// BlockStatement represents a block of statements
type BlockStatement struct {
	Token      lexer.Token // The '{' token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out strings.Builder
	out.WriteString("{\n")
	for _, s := range bs.Statements {
		out.WriteString("  " + s.String() + "\n")
	}
	out.WriteString("}")
	return out.String()
}

// IfStatement represents an if statement
type IfStatement struct {
	Token       lexer.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative Statement // Can be another IfStatement or BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	out := "if " + is.Condition.String() + " " + is.Consequence.String()
	if is.Alternative != nil {
		out += " else " + is.Alternative.String()
	}
	return out
}

// ForStatement represents a for loop
type ForStatement struct {
	Token     lexer.Token // The 'for' token
	Init      Statement   // Optional initialization
	Condition Expression  // Loop condition
	Post      Statement   // Optional post statement
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	out := "for "
	if fs.Init != nil {
		out += fs.Init.String() + " "
	}
	out += fs.Condition.String()
	if fs.Post != nil {
		out += "; " + fs.Post.String()
	}
	out += " " + fs.Body.String()
	return out
}

// ReturnStatement represents a return statement
type ReturnStatement struct {
	Token       lexer.Token // The 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	out := "return"
	if rs.ReturnValue != nil {
		out += " " + rs.ReturnValue.String()
	}
	return out + ";"
}

// BreakStatement represents a break statement
type BreakStatement struct {
	Token lexer.Token // The 'break' token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return "break;" }

// ContinueStatement represents a continue statement
type ContinueStatement struct {
	Token lexer.Token // The 'continue' token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return "continue;" }

// ExpressionStatement represents an expression as a statement
type ExpressionStatement struct {
	Token      lexer.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String() + ";"
	}
	return ""
}

// FunctionParameter represents a function parameter
type FunctionParameter struct {
	Name *Identifier
	Type *TypeAnnotation
}

// FunctionStatement represents a function declaration
type FunctionStatement struct {
	Token      lexer.Token // The 'func' token
	Name       *Identifier
	Parameters []*FunctionParameter
	ReturnType *TypeAnnotation
	Body       *BlockStatement
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) String() string {
	var params []string
	for _, p := range fs.Parameters {
		params = append(params, p.Name.String()+": "+p.Type.String())
	}
	out := "func " + fs.Name.String() + "(" + strings.Join(params, ", ") + ")"
	if fs.ReturnType != nil {
		out += ": " + fs.ReturnType.String()
	}
	out += " " + fs.Body.String()
	return out
}

// StructField represents a struct field
type StructField struct {
	Name *Identifier
	Type *TypeAnnotation
}

// TypeStatement represents a type definition
type TypeStatement struct {
	Token      lexer.Token // The 'type' token
	Name       *Identifier
	Definition Statement // StructStatement or EnumStatement
}

func (ts *TypeStatement) statementNode()       {}
func (ts *TypeStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *TypeStatement) String() string {
	return "type " + ts.Name.String() + " = " + ts.Definition.String()
}

// StructStatement represents a struct declaration
type StructStatement struct {
	Token  lexer.Token // The 'struct' token
	Name   *Identifier
	Fields []*StructField
}

func (ss *StructStatement) statementNode()       {}
func (ss *StructStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *StructStatement) String() string {
	var fields []string
	for _, f := range ss.Fields {
		fields = append(fields, f.Name.String()+": "+f.Type.String())
	}
	return "struct " + ss.Name.String() + " {\n  " + strings.Join(fields, ";\n  ") + ";\n}"
}

// EnumStatement represents an enum declaration
type EnumStatement struct {
	Token    lexer.Token   // The 'enum' token
	Name     *Identifier
	Variants []*Identifier // List of enum variant names
}

func (es *EnumStatement) statementNode()       {}
func (es *EnumStatement) TokenLiteral() string { return es.Token.Literal }
func (es *EnumStatement) String() string {
	var variants []string
	for _, v := range es.Variants {
		variants = append(variants, v.String())
	}
	return "enum " + es.Name.String() + " { " + strings.Join(variants, ", ") + " }"
}

// SwitchStatement represents a switch statement
type SwitchStatement struct {
	Token   lexer.Token // The 'switch' token
	Value   Expression  // The value being switched on
	Cases   []*CaseClause
	Default *BlockStatement // Optional default case
}

func (ss *SwitchStatement) statementNode()       {}
func (ss *SwitchStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SwitchStatement) String() string {
	out := "switch " + ss.Value.String() + " {\n"
	for _, c := range ss.Cases {
		out += c.String() + "\n"
	}
	if ss.Default != nil {
		out += "default " + ss.Default.String() + "\n"
	}
	out += "}"
	return out
}

// CaseClause represents a case in a switch statement
type CaseClause struct {
	Token lexer.Token // The 'case' token
	Value Expression  // The value to match
	Body  *BlockStatement
}

func (cc *CaseClause) String() string {
	return "case " + cc.Value.String() + " " + cc.Body.String()
}
