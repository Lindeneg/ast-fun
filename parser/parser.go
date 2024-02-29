package parser

import (
	"fmt"
	"strconv"

	"github.com/lindeneg/monkey/ast"
	"github.com/lindeneg/monkey/lexer"
	"github.com/lindeneg/monkey/token"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)

// Define the precedence of the different token types
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
}

type Parser struct {
	l              *lexer.Lexer
	current        token.Token
	next           token.Token
	errors         []string
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	// TODO can probably use a global map for this
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	// Read two tokens, so current and next are both set
	p.nextToken()
	p.nextToken()
	return p
}

// Parse the program and return an ast.Program
// The parser will keep parsing until it encounters an EOF token
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}
	for p.current.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}
	return program
}

// Return the errors encountered during parsing
func (p *Parser) Errors() []string {
	return p.errors
}

// Parse an identifier and return an ast.Identifier
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.current, Value: p.current.Literal}
}

// Find the precedence of the current token
// If the current token is not in the map, return LOWEST
func (p *Parser) currentPrecedence() int {
	if p, ok := precedences[p.current.Type]; ok {
		return p
	}
	return LOWEST
}

// Find the precedence of the next token
// If the next token is not in the map, return LOWEST
func (p *Parser) nextPrecedence() int {
	if p, ok := precedences[p.next.Type]; ok {
		return p
	}
	return LOWEST
}

// Move the current token to the next token
func (p *Parser) nextToken() {
	p.current = p.next
	p.next = p.l.NextToken()
}

// Parse a statement and return an ast.Statement
func (p *Parser) parseStatement() ast.Statement {
	switch p.current.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.current, Value: p.currentTokenIs(token.TRUE)}
}

// Parse a let statement and return an ast.LetStatement
func (p *Parser) parseLetStatement() *ast.LetStatement {
	defer untrace(trace("parseLetStatement"))

	stmt := &ast.LetStatement{Token: p.current}
	if !p.expectNext(token.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.current, Value: p.current.Literal}
	if !p.expectNext(token.ASSIGN) {
		return nil
	}
	// TODO: we are skipping expressions until ; is encoutnered
	for !p.currentTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// Parse a return statement and return an ast.ReturnStatement
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	defer untrace(trace("parseReturnStatement"))

	stmt := &ast.ReturnStatement{Token: p.current}
	p.nextToken()

	for !p.currentTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// Parse an expression statement and return an ast.ExpressionStatement
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer untrace(trace("parseExpressionStatement"))

	stmt := &ast.ExpressionStatement{Token: p.current}
	stmt.Expression = p.parseExpression(LOWEST)
	if p.nextTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

// Parse a prefix expression and return an ast.PrefixExpression
func (p *Parser) parsePrefixExpression() ast.Expression {
	defer untrace(trace("parsePrefixExpression"))

	expression := &ast.PrefixExpression{
		Token:    p.current,
		Operator: p.current.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))

	expression := &ast.InfixExpression{
		Token:    p.current,
		Operator: p.current.Literal,
		Left:     left,
	}
	precedence := p.currentPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

// Parse an expression and return an ast.Expression
func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer untrace(trace("parseExpression"))

	prefix := p.prefixParseFns[p.current.Type]
	if prefix == nil {
		msg := fmt.Sprintf("no prefix parse function for %s found", p.current.Type)
		p.errors = append(p.errors, msg)
		return nil
	}
	lhs := prefix()

	for !p.nextTokenIs(token.SEMICOLON) && precedence < p.nextPrecedence() {
		infix := p.infixParseFns[p.next.Type]
		if infix == nil {
			return lhs
		}
		p.nextToken()
		lhs = infix(lhs)
	}
	return lhs
}

// Parse an integer literal and return an ast.Expression
func (p *Parser) parseIntegerLiteral() ast.Expression {
	defer untrace(trace("parseIntegerLiteral"))

	lit := &ast.IntegerLiteral{Token: p.current}
	value, err := strconv.ParseInt(p.current.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.current.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}
	lit.Value = value
	return lit
}

// test
func (p *Parser) currentTokenIs(t token.TokenType) bool {
	return p.current.Type == t
}

func (p *Parser) nextTokenIs(t token.TokenType) bool {
	return p.next.Type == t
}

func (p *Parser) expectNext(t token.TokenType) bool {
	if p.nextTokenIs(t) {
		p.nextToken()
		return true
	}
	msg := fmt.Sprintf("expected next token to be %s but got %s instead",
		t, p.next.Type)
	p.errors = append(p.errors, msg)
	return false
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
