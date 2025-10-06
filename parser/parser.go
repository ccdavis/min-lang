package parser

import (
	"fmt"
	"minlang/ast"
	"minlang/lexer"
	"strconv"
)

// Precedence levels for operators
const (
	_ int = iota
	LOWEST
	OR          // ||
	AND         // &&
	EQUALS      // ==, !=
	LESSGREATER // <, >, <=, >=
	SUM         // +, -
	PRODUCT     // *, /, %
	PREFIX      // -x, !x
	CALL        // func(x), x[y], x.y
)

var precedences = map[lexer.TokenType]int{
	lexer.OR:       OR,
	lexer.AND:      AND,
	lexer.EQ:       EQUALS,
	lexer.NE:       EQUALS,
	lexer.LT:       LESSGREATER,
	lexer.GT:       LESSGREATER,
	lexer.LE:       LESSGREATER,
	lexer.GE:       LESSGREATER,
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.ASTERISK: PRODUCT,
	lexer.SLASH:    PRODUCT,
	lexer.PERCENT:  PRODUCT,
	lexer.LPAREN:   CALL,
	lexer.LBRACKET: CALL,
	lexer.DOT:      CALL,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser represents the parser
type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

// New creates a new parser
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Initialize prefix parse functions
	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.NIL, p.parseNilLiteral)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(lexer.MAP, p.parseMapLiteral)

	// Initialize infix parse functions
	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.ASTERISK, p.parseInfixExpression)
	p.registerInfix(lexer.SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.PERCENT, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NE, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.GT, p.parseInfixExpression)
	p.registerInfix(lexer.LE, p.parseInfixExpression)
	p.registerInfix(lexer.GE, p.parseInfixExpression)
	p.registerInfix(lexer.AND, p.parseInfixExpression)
	p.registerInfix(lexer.OR, p.parseInfixExpression)
	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.LBRACKET, p.parseIndexExpression)
	p.registerInfix(lexer.DOT, p.parseFieldAccessExpression)

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns the parser errors
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead at line %d, column %d",
		t, p.peekToken.Type, p.peekToken.Line, p.peekToken.Column)
	p.errors = append(p.errors, msg)
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// ParseProgram parses the entire program
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case lexer.VAR:
		return p.parseVarStatement(true)
	case lexer.CONST:
		return p.parseVarStatement(false)
	case lexer.FUNC:
		return p.parseFunctionStatement()
	case lexer.TYPE:
		return p.parseTypeStatement()
	case lexer.STRUCT:
		return p.parseStructStatement()
	case lexer.ENUM:
		return p.parseEnumStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.BREAK:
		return p.parseBreakStatement()
	case lexer.CONTINUE:
		return p.parseContinueStatement()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.SWITCH:
		return p.parseSwitchStatement()
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.LBRACE:
		return p.parseBlockStatement()
	default:
		// Try to parse as assignment or expression statement
		return p.parseExpressionOrAssignmentStatement()
	}
}

func (p *Parser) parseVarStatement(isMutable bool) *ast.VarStatement {
	stmt := &ast.VarStatement{Token: p.curToken, IsMutable: isMutable}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Optional type annotation
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // consume ':'
		p.nextToken() // move to type
		stmt.Type = p.parseTypeAnnotation()
	}

	// Optional initialization
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // consume '='
		p.nextToken() // move to value
		stmt.Value = p.parseAssignmentValue()
	}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseTypeAnnotation() *ast.TypeAnnotation {
	ta := &ast.TypeAnnotation{Token: p.curToken}

	// Check for array type
	if p.curTokenIs(lexer.LBRACKET) {
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}
		ta.IsArray = true
		p.nextToken() // move to element type
		ta.ElementType = p.parseTypeAnnotation()
		return ta
	}

	// Check for map type
	if p.curTokenIs(lexer.MAP) {
		ta.IsMap = true
		if !p.expectPeek(lexer.LBRACKET) {
			return nil
		}
		p.nextToken() // move to key type
		ta.KeyType = p.parseTypeAnnotation()
		if !p.expectPeek(lexer.RBRACKET) {
			return nil
		}
		p.nextToken() // move to value type
		ta.ValueType = p.parseTypeAnnotation()
		return ta
	}

	// Simple type (identifier)
	if p.curTokenIs(lexer.IDENT) {
		ta.Name = p.curToken.Literal
		return ta
	}

	return nil
}

func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	stmt := &ast.FunctionStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	// Optional return type
	if p.peekTokenIs(lexer.COLON) {
		p.nextToken() // consume ':'
		p.nextToken() // move to return type
		stmt.ReturnType = p.parseTypeAnnotation()
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseFunctionParameters() []*ast.FunctionParameter {
	params := []*ast.FunctionParameter{}

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken() // move to first parameter

	param := &ast.FunctionParameter{}
	param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken() // move to type
	param.Type = p.parseTypeAnnotation()
	params = append(params, param)

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume ','
		p.nextToken() // move to next parameter

		param := &ast.FunctionParameter{}
		param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken() // move to type
		param.Type = p.parseTypeAnnotation()
		params = append(params, param)
	}

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return params
}

func (p *Parser) parseTypeStatement() *ast.TypeStatement {
	stmt := &ast.TypeStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.ASSIGN) {
		return nil
	}

	p.nextToken() // move to struct or enum

	switch p.curToken.Type {
	case lexer.STRUCT:
		stmt.Definition = p.parseStructDefinition()
	case lexer.ENUM:
		stmt.Definition = p.parseEnumDefinition()
	default:
		msg := fmt.Sprintf("expected struct or enum after =, got %s", p.curToken.Type.String())
		p.errors = append(p.errors, msg)
		return nil
	}

	return stmt
}

func (p *Parser) parseStructDefinition() *ast.StructStatement {
	stmt := &ast.StructStatement{Token: p.curToken}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Fields = p.parseStructFields()

	return stmt
}

func (p *Parser) parseStructStatement() *ast.StructStatement {
	stmt := &ast.StructStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Fields = p.parseStructFields()

	return stmt
}

func (p *Parser) parseStructFields() []*ast.StructField {
	fields := []*ast.StructField{}

	p.nextToken() // move to first field or '}'

	if p.curTokenIs(lexer.RBRACE) {
		return fields
	}

	field := &ast.StructField{}
	field.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken() // move to type
	field.Type = p.parseTypeAnnotation()
	fields = append(fields, field)

	// Support both semicolon and comma as field separators
	if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
	}

	for !p.peekTokenIs(lexer.RBRACE) {
		p.nextToken() // move to next field

		field := &ast.StructField{}
		field.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken() // move to type
		field.Type = p.parseTypeAnnotation()
		fields = append(fields, field)

		// Support both semicolon and comma as field separators
		if p.peekTokenIs(lexer.SEMICOLON) || p.peekTokenIs(lexer.COMMA) {
			p.nextToken()
		}
	}

	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}

	return fields
}

func (p *Parser) parseEnumDefinition() *ast.EnumStatement {
	stmt := &ast.EnumStatement{Token: p.curToken}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Variants = p.parseEnumVariants()

	return stmt
}

func (p *Parser) parseEnumStatement() *ast.EnumStatement {
	stmt := &ast.EnumStatement{Token: p.curToken}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Variants = p.parseEnumVariants()

	return stmt
}

func (p *Parser) parseEnumVariants() []*ast.Identifier {
	variants := []*ast.Identifier{}

	p.nextToken() // move to first variant or '}'

	if p.curTokenIs(lexer.RBRACE) {
		return variants
	}

	// First variant
	variant := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	variants = append(variants, variant)

	// Additional variants
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume ','
		p.nextToken() // move to next variant

		variant := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		variants = append(variants, variant)
	}

	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}

	return variants
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	if !p.curTokenIs(lexer.SEMICOLON) {
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{Token: p.curToken}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	stmt := &ast.ContinueStatement{Token: p.curToken}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	p.nextToken() // move to condition
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(lexer.ELSE) {
		p.nextToken() // consume 'else'

		if p.peekTokenIs(lexer.IF) {
			p.nextToken() // consume 'if'
			stmt.Alternative = p.parseIfStatement()
		} else if p.peekTokenIs(lexer.LBRACE) {
			p.nextToken() // consume '{'
			stmt.Alternative = p.parseBlockStatement()
		}
	}

	return stmt
}

func (p *Parser) parseSwitchStatement() *ast.SwitchStatement {
	stmt := &ast.SwitchStatement{Token: p.curToken}

	p.nextToken() // move to switch value

	// Parse switch value - can be identifier, integer, or enum value
	// Don't use full parseExpression to avoid struct literal ambiguity
	switch p.curToken.Type {
	case lexer.IDENT:
		stmt.Value = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	case lexer.INT:
		value, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
		stmt.Value = &ast.IntegerLiteral{Token: p.curToken, Value: value}
	default:
		msg := fmt.Sprintf("expected identifier or integer in switch, got %s", p.curToken.Type.String())
		p.errors = append(p.errors, msg)
		return nil
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	p.nextToken() // move to first case or default or '}'

	// Parse case clauses
	for p.curTokenIs(lexer.CASE) {
		caseClause := &ast.CaseClause{Token: p.curToken}

		p.nextToken() // move to case value
		// Parse case value - can be identifier or integer
		// Don't use full parseExpression to avoid struct literal ambiguity
		switch p.curToken.Type {
		case lexer.IDENT:
			caseClause.Value = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		case lexer.INT:
			value, _ := strconv.ParseInt(p.curToken.Literal, 10, 64)
			caseClause.Value = &ast.IntegerLiteral{Token: p.curToken, Value: value}
		default:
			msg := fmt.Sprintf("expected identifier or integer in case, got %s", p.curToken.Type.String())
			p.errors = append(p.errors, msg)
			return nil
		}

		if !p.expectPeek(lexer.LBRACE) {
			return nil
		}

		caseClause.Body = p.parseBlockStatement()
		stmt.Cases = append(stmt.Cases, caseClause)

		p.nextToken() // move to next case, default, or '}'
	}

	// Parse optional default clause
	if p.curTokenIs(lexer.DEFAULT) {
		if !p.expectPeek(lexer.LBRACE) {
			return nil
		}

		stmt.Default = p.parseBlockStatement()
		p.nextToken() // move past '}'
	}

	// Expect closing brace
	if !p.curTokenIs(lexer.RBRACE) {
		msg := fmt.Sprintf("expected }, got %s instead", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	p.nextToken() // move past 'for'

	// Simple for loop: for condition { ... }
	if !p.curTokenIs(lexer.VAR) && !p.curTokenIs(lexer.CONST) {
		stmt.Condition = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.LBRACE) {
			return nil
		}
		stmt.Body = p.parseBlockStatement()
		return stmt
	}

	// C-style for loop: for init; condition; post { ... }
	stmt.Init = p.parseVarStatement(true)

	// The var statement should have consumed the semicolon
	// Now parse the condition
	p.nextToken() // move to condition
	stmt.Condition = p.parseExpression(LOWEST)

	if p.expectPeek(lexer.SEMICOLON) {
		p.nextToken() // move to post statement
		stmt.Post = p.parseExpressionOrAssignmentStatement()
	}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(lexer.RBRACE) && !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseExpressionOrAssignmentStatement() ast.Statement {
	// Try to parse the left side
	expr := p.parseExpression(LOWEST)

	// Check if this is an assignment
	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // consume '='
		stmt := &ast.AssignmentStatement{
			Token: p.curToken,
			Left:  expr,
		}
		p.nextToken() // move to value
		stmt.Value = p.parseAssignmentValue()

		if p.peekTokenIs(lexer.SEMICOLON) {
			p.nextToken()
		}

		return stmt
	}

	// It's an expression statement
	stmt := &ast.ExpressionStatement{
		Token:      p.curToken,
		Expression: expr,
	}

	if p.peekTokenIs(lexer.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseAssignmentValue parses a value in assignment context, allowing struct literals
func (p *Parser) parseAssignmentValue() ast.Expression {
	// Check if this is a struct literal: Identifier {
	if p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.LBRACE) {
		return p.parseStructLiteral()
	}
	// Otherwise parse as normal expression
	return p.parseExpression(LOWEST)
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found at line %d, column %d",
		t, p.curToken.Line, p.curToken.Column)
	p.errors = append(p.errors, msg)
}

// Expression parsing functions

func (p *Parser) parseIdentifier() ast.Expression {
	// Don't try to parse struct literals here to avoid ambiguity
	// Struct literals will be handled in specific contexts (assignments)
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(lexer.TRUE)}
}

func (p *Parser) parseNilLiteral() ast.Expression {
	return &ast.NilLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(lexer.RPAREN)
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseFieldAccessExpression(left ast.Expression) ast.Expression {
	exp := &ast.FieldAccessExpression{Token: p.curToken, Left: left}

	if !p.expectPeek(lexer.IDENT) {
		return nil
	}

	exp.Field = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return exp
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(lexer.RBRACKET)
	return array
}

func (p *Parser) parseMapLiteral() ast.Expression {
	mapLit := &ast.MapLiteral{Token: p.curToken}

	if !p.expectPeek(lexer.LBRACKET) {
		return nil
	}

	p.nextToken() // move to key type
	mapLit.KeyType = p.parseTypeAnnotation()

	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}

	p.nextToken() // move to value type
	mapLit.ValueType = p.parseTypeAnnotation()

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	mapLit.Pairs = p.parseMapPairs()

	return mapLit
}

func (p *Parser) parseMapPairs() map[ast.Expression]ast.Expression {
	pairs := make(map[ast.Expression]ast.Expression)

	if p.peekTokenIs(lexer.RBRACE) {
		p.nextToken()
		return pairs
	}

	p.nextToken() // move to first key
	key := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken() // move to value
	value := p.parseExpression(LOWEST)

	pairs[key] = value

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume ','
		p.nextToken() // move to next key

		key := p.parseExpression(LOWEST)

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken() // move to value
		value := p.parseExpression(LOWEST)

		pairs[key] = value
	}

	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}

	return pairs
}

func (p *Parser) parseStructLiteral() ast.Expression {
	structLit := &ast.StructLiteral{Token: p.curToken}
	structLit.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LBRACE) {
		return nil
	}

	structLit.Fields = p.parseStructLiteralFields()

	return structLit
}

func (p *Parser) parseStructLiteralFields() map[string]ast.Expression {
	fields := make(map[string]ast.Expression)

	if p.peekTokenIs(lexer.RBRACE) {
		p.nextToken()
		return fields
	}

	p.nextToken() // move to first field name

	fieldName := p.curToken.Literal

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	p.nextToken() // move to value
	value := p.parseExpression(LOWEST)

	fields[fieldName] = value

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume ','
		p.nextToken() // move to next field name

		fieldName := p.curToken.Literal

		if !p.expectPeek(lexer.COLON) {
			return nil
		}

		p.nextToken() // move to value
		value := p.parseExpression(LOWEST)

		fields[fieldName] = value
	}

	if !p.expectPeek(lexer.RBRACE) {
		return nil
	}

	return fields
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken() // consume ','
		p.nextToken() // move to next expression
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}
