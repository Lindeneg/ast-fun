package lexer

import (
	"github.com/lindeneg/monkey/token"
)

type Lexer struct {
	input string
	// current index
	current int
	// next index
	next int
	// current char
	char byte
	// current line
	Lines int
	// current col
	Col int
}

// Creates a new lexer and reads the
// first character of the input string
func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, Lines: 1, Col: 1}
	l.readChar()
	return l
}

// Returns the next token in the input string
// and advances the position in the input string
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.ignoreWhitespace()
	tok.Line = l.Lines
	tok.Col = l.Col
	switch l.char {
	case '=':
		if l.peekChar() == '=' {
			tok = tokenWithNext(l, token.EQ)
		} else {
			tok = l.newToken(token.ASSIGN, l.char)
		}
	case ';':
		tok = l.newToken(token.SEMICOLON, l.char)
	case '(':
		tok = l.newToken(token.LPAREN, l.char)
	case ')':
		tok = l.newToken(token.RPAREN, l.char)
	case ',':
		tok = l.newToken(token.COMMA, l.char)
	case '+':
		tok = l.newToken(token.PLUS, l.char)
	case '-':
		tok = l.newToken(token.MINUS, l.char)
	case '[':
		tok = l.newToken(token.LBRACKET, l.char)
	case ']':
		tok = l.newToken(token.RBRACKET, l.char)
	case ':':
		tok = l.newToken(token.COLON, l.char)
	case '!':
		if l.peekChar() == '=' {
			tok = tokenWithNext(l, token.NOT_EQ)
		} else {
			tok = l.newToken(token.BANG, l.char)
		}
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
	case '/':
		if l.peekChar() == '/' {
			l.ignoreComment()
			return l.NextToken()
		} else {
			tok = l.newToken(token.SLASH, l.char)
		}
	case '*':
		tok = l.newToken(token.ASTERISK, l.char)
	case '<':
		if l.peekChar() == '=' {
			tok = tokenWithNext(l, token.LT_OR_EQ)
		} else {
			tok = l.newToken(token.LT, l.char)
		}
	case '>':
		if l.peekChar() == '=' {
			tok = tokenWithNext(l, token.GT_OR_EQ)
		} else {
			tok = l.newToken(token.GT, l.char)
		}
	case '{':
		tok = l.newToken(token.LBRACE, l.char)
	case '}':
		tok = l.newToken(token.RBRACE, l.char)
	default:
		if isIdentifierByte(l.char) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			// return early to avoid readChar() call
			// readIdentifier() has already advanced position
			return tok
		} else if isDigit(l.char) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		}
		tok = l.newToken(token.ILLEGAL, l.char)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	}
	l.readChar()
	return tok
}

func (l *Lexer) readString() string {
	position := l.current + 1
	for {
		l.readChar()
		if l.char == '"' || l.char == 0 {
			break
		}
	}
	return l.input[position:l.current]
}

func (l *Lexer) ignoreComment() {
	for l.char != 0 && l.char != '\n' {
		l.readChar()
	}
}

// read without incrementing the next position
func (l *Lexer) peekChar() byte {
	if l.next >= len(l.input) {
		return 0
	}
	return l.input[l.next]
}

// read the next character and advance the position
func (l *Lexer) readChar() {
	if l.char == '\n' {
		l.Lines += 1
		l.Col = 1
	} else {
		l.Col += 1
	}
	if l.next >= len(l.input) {
		l.char = 0
	} else {
		l.char = l.input[l.next]
	}
	l.current = l.next
	l.next += 1
}

// read a number from current pos in input string
func (l *Lexer) readNumber() string {
	return readUntil(l, isDigit)
}

// read an identifier from current pos in input string
func (l *Lexer) readIdentifier() string {
	return readUntil(l, isIdentifierByte)
}

// ignore whitespace characters
func (l *Lexer) ignoreWhitespace() {
	for l.char == ' ' || l.char == '\t' || l.char == '\n' || l.char == '\r' {
		l.readChar()
	}
}

// return a token with the next character appended to the literal
// and advance the position accordingly
func tokenWithNext(l *Lexer, tokenType token.TokenType) token.Token {
	return tokenFromRange(l, tokenType, 1)
}

// return a token with the next r characters appended to the literal
// and advance the position accordingly
func tokenFromRange(l *Lexer, tokenType token.TokenType, r int) token.Token {
	literal := string(l.char)
	for i := 0; i < r; i++ {
		l.readChar()
		literal += string(l.char)
	}
	return token.Token{Type: tokenType, Literal: literal}
}

type readUntilCallback func(char byte) bool

// read until the callback returns false
// and return the string read
// and advance the position accordingly
func readUntil(l *Lexer, shouldContinue readUntilCallback) string {
	pos := l.current
	for shouldContinue(l.char) {
		l.readChar()
	}
	return l.input[pos:l.current]
}

// This checks if char is a valid identifer byte.
// If more characters should be supported,
// they should be added here in the conditional
func isIdentifierByte(char byte) bool {
	return char >= 'a' && char <= 'z' ||
		char >= 'A' && char <= 'Z' ||
		char == '_'
}

func isDigit(char byte) bool {
	return char >= '0' && char <= '9'
}

func (l *Lexer) newToken(tokenType token.TokenType, char byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(char), Line: l.Lines, Col: l.Col}
}
