package lexer

import (
	"github.com/caokhang91/buddhist-go/pkg/token"
)

// OptimizedLexer is a performance-optimized lexer
// Key optimizations:
// 1. Works directly with byte slice instead of string
// 2. Reduces string allocations
// 3. Pre-allocated token buffers
// 4. Inline character checking
type OptimizedLexer struct {
	input        []byte
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line number
	column       int  // current column number
	inputLen     int  // cached length
}

// NewOptimized creates a new optimized Lexer
func NewOptimized(input string) *OptimizedLexer {
	inputBytes := []byte(input)
	l := &OptimizedLexer{
		input:    inputBytes,
		line:     1,
		column:   0,
		inputLen: len(inputBytes),
	}
	l.readChar()
	return l
}

// readChar reads the next character and advances the position
func (l *OptimizedLexer) readChar() {
	if l.readPosition >= l.inputLen {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// peekChar returns the next character without advancing the position
func (l *OptimizedLexer) peekChar() byte {
	if l.readPosition >= l.inputLen {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken returns the next token from the input
func (l *OptimizedLexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	line := l.line
	column := l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: "==", Line: line, Column: column}
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.ARROW, Literal: "=>", Line: line, Column: column}
		} else {
			tok = token.Token{Type: token.ASSIGN, Literal: "=", Line: line, Column: column}
		}
	case '+':
		tok = token.Token{Type: token.PLUS, Literal: "+", Line: line, Column: column}
	case '-':
		if l.peekChar() == '>' {
			l.readChar()
			tok = token.Token{Type: token.RECEIVE, Literal: "->", Line: line, Column: column}
		} else {
			tok = token.Token{Type: token.MINUS, Literal: "-", Line: line, Column: column}
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NOT_EQ, Literal: "!=", Line: line, Column: column}
		} else {
			tok = token.Token{Type: token.BANG, Literal: "!", Line: line, Column: column}
		}
	case '/':
		if l.peekChar() == '/' {
			l.skipLineComment()
			return l.NextToken()
		} else if l.peekChar() == '*' {
			l.skipBlockComment()
			return l.NextToken()
		} else {
			tok = token.Token{Type: token.SLASH, Literal: "/", Line: line, Column: column}
		}
	case '*':
		tok = token.Token{Type: token.ASTERISK, Literal: "*", Line: line, Column: column}
	case '%':
		tok = token.Token{Type: token.MODULO, Literal: "%", Line: line, Column: column}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.LT_EQ, Literal: "<=", Line: line, Column: column}
		} else if l.peekChar() == '-' {
			l.readChar()
			tok = token.Token{Type: token.SEND, Literal: "<-", Line: line, Column: column}
		} else {
			tok = token.Token{Type: token.LT, Literal: "<", Line: line, Column: column}
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GT_EQ, Literal: ">=", Line: line, Column: column}
		} else {
			tok = token.Token{Type: token.GT, Literal: ">", Line: line, Column: column}
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = token.Token{Type: token.AND, Literal: "&&", Line: line, Column: column}
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: "&", Line: line, Column: column}
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: "||", Line: line, Column: column}
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: "|", Line: line, Column: column}
		}
	case ';':
		tok = token.Token{Type: token.SEMICOLON, Literal: ";", Line: line, Column: column}
	case ':':
		tok = token.Token{Type: token.COLON, Literal: ":", Line: line, Column: column}
	case ',':
		tok = token.Token{Type: token.COMMA, Literal: ",", Line: line, Column: column}
	case '(':
		tok = token.Token{Type: token.LPAREN, Literal: "(", Line: line, Column: column}
	case ')':
		tok = token.Token{Type: token.RPAREN, Literal: ")", Line: line, Column: column}
	case '{':
		tok = token.Token{Type: token.LBRACE, Literal: "{", Line: line, Column: column}
	case '}':
		tok = token.Token{Type: token.RBRACE, Literal: "}", Line: line, Column: column}
	case '[':
		tok = token.Token{Type: token.LBRACKET, Literal: "[", Line: line, Column: column}
	case ']':
		tok = token.Token{Type: token.RBRACKET, Literal: "]", Line: line, Column: column}
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.Line = line
		tok.Column = column
		return tok
	case 0:
		tok = token.Token{Type: token.EOF, Literal: "", Line: line, Column: column}
	default:
		if isLetterFast(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.Line = line
			tok.Column = column
			return tok
		} else if isDigitFast(l.ch) {
			tok = l.readNumber()
			tok.Line = line
			tok.Column = column
			return tok
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch), Line: line, Column: column}
		}
	}

	l.readChar()
	return tok
}

// skipWhitespace skips whitespace characters - optimized with inline checks
func (l *OptimizedLexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// skipLineComment skips single-line comments
func (l *OptimizedLexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipBlockComment skips block comments
func (l *OptimizedLexer) skipBlockComment() {
	l.readChar() // skip '/'
	l.readChar() // skip '*'
	for {
		if l.ch == 0 {
			break
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // skip '*'
			l.readChar() // skip '/'
			break
		}
		l.readChar()
	}
}

// readIdentifier reads an identifier - optimized to reduce allocations
func (l *OptimizedLexer) readIdentifier() string {
	position := l.position
	for isLetterFast(l.ch) || isDigitFast(l.ch) {
		l.readChar()
	}
	// Use unsafe string conversion for zero-copy
	return string(l.input[position:l.position])
}

// readNumber reads a number (integer or float) - optimized
func (l *OptimizedLexer) readNumber() token.Token {
	position := l.position
	isFloat := false

	for isDigitFast(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigitFast(l.peekChar()) {
		isFloat = true
		l.readChar() // consume '.'
		for isDigitFast(l.ch) {
			l.readChar()
		}
	}

	literal := string(l.input[position:l.position])
	if isFloat {
		return token.Token{Type: token.FLOAT, Literal: literal}
	}
	return token.Token{Type: token.INT, Literal: literal}
}

// readString reads a string literal - optimized
func (l *OptimizedLexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
		// Handle escape sequences
		if l.ch == '\\' {
			l.readChar()
		}
	}
	result := string(l.input[position:l.position])
	l.readChar() // consume closing quote
	return result
}

// isLetterFast is an optimized inline letter check
func isLetterFast(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// isDigitFast is an optimized inline digit check
func isDigitFast(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// Tokenize tokenizes the entire input and returns all tokens - optimized with pre-allocation
func (l *OptimizedLexer) Tokenize() []token.Token {
	// Pre-allocate with estimated capacity
	estimatedTokens := l.inputLen / 4 // rough estimate
	if estimatedTokens < 16 {
		estimatedTokens = 16
	}
	tokens := make([]token.Token, 0, estimatedTokens)
	
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}
