package lexer

import (
	"github.com/caokhang91/buddhist-go/pkg/token"
)

// TokenLexer is an interface that both Lexer and OptimizedLexer implement
type TokenLexer interface {
	NextToken() token.Token
}

// Lexer represents the lexical analyzer
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line number
	column       int  // current column number
}

// New creates a new Lexer
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

// readChar reads the next character and advances the position
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
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
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// NextToken returns the next token from the input
func (l *Lexer) NextToken() token.Token {
	var tok token.Token

	l.skipWhitespace()

	line := l.line
	column := l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.New(token.EQ, "==", line, column)
		} else if l.peekChar() == '>' {
			l.readChar()
			tok = token.New(token.ARROW, "=>", line, column)
		} else {
			tok = token.New(token.ASSIGN, string(l.ch), line, column)
		}
	case '+':
		tok = token.New(token.PLUS, string(l.ch), line, column)
	case '-':
		if l.peekChar() == '>' {
			l.readChar()
			tok = token.New(token.RECEIVE, "->", line, column)
		} else {
			tok = token.New(token.MINUS, string(l.ch), line, column)
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.New(token.NOT_EQ, "!=", line, column)
		} else {
			tok = token.New(token.BANG, string(l.ch), line, column)
		}
	case '/':
		if l.peekChar() == '/' {
			l.skipLineComment()
			return l.NextToken()
		} else if l.peekChar() == '*' {
			l.skipBlockComment()
			return l.NextToken()
		} else {
			tok = token.New(token.SLASH, string(l.ch), line, column)
		}
	case '*':
		tok = token.New(token.ASTERISK, string(l.ch), line, column)
	case '%':
		tok = token.New(token.MODULO, string(l.ch), line, column)
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.New(token.LT_EQ, "<=", line, column)
		} else if l.peekChar() == '-' {
			l.readChar()
			tok = token.New(token.SEND, "<-", line, column)
		} else {
			tok = token.New(token.LT, string(l.ch), line, column)
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.New(token.GT_EQ, ">=", line, column)
		} else {
			tok = token.New(token.GT, string(l.ch), line, column)
		}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = token.New(token.AND, "&&", line, column)
		} else {
			tok = token.New(token.ILLEGAL, string(l.ch), line, column)
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = token.New(token.OR, "||", line, column)
		} else {
			tok = token.New(token.ILLEGAL, string(l.ch), line, column)
		}
	case ';':
		tok = token.New(token.SEMICOLON, string(l.ch), line, column)
	case ':':
		tok = token.New(token.COLON, string(l.ch), line, column)
	case ',':
		tok = token.New(token.COMMA, string(l.ch), line, column)
	case '(':
		tok = token.New(token.LPAREN, string(l.ch), line, column)
	case ')':
		tok = token.New(token.RPAREN, string(l.ch), line, column)
	case '{':
		tok = token.New(token.LBRACE, string(l.ch), line, column)
	case '}':
		tok = token.New(token.RBRACE, string(l.ch), line, column)
	case '[':
		tok = token.New(token.LBRACKET, string(l.ch), line, column)
	case ']':
		tok = token.New(token.RBRACKET, string(l.ch), line, column)
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.Line = line
		tok.Column = column
		return tok
	case 0:
		tok = token.New(token.EOF, "", line, column)
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.Line = line
			tok.Column = column
			return tok
		} else if isDigit(l.ch) {
			tok = l.readNumber()
			tok.Line = line
			tok.Column = column
			return tok
		} else {
			tok = token.New(token.ILLEGAL, string(l.ch), line, column)
		}
	}

	l.readChar()
	return tok
}

// skipWhitespace skips whitespace characters
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// skipLineComment skips single-line comments
func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipBlockComment skips block comments
func (l *Lexer) skipBlockComment() {
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

// readIdentifier reads an identifier
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a number (integer or float)
func (l *Lexer) readNumber() token.Token {
	position := l.position
	isFloat := false

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.input[position:l.position]
	if isFloat {
		return token.Token{Type: token.FLOAT, Literal: literal}
	}
	return token.Token{Type: token.INT, Literal: literal}
}

// readString reads a string literal
func (l *Lexer) readString() string {
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
	result := l.input[position:l.position]
	l.readChar() // consume closing quote
	return result
}

// isLetter checks if the character is a letter
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit checks if the character is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// Tokenize tokenizes the entire input and returns all tokens
func (l *Lexer) Tokenize() []token.Token {
	var tokens []token.Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == token.EOF {
			break
		}
	}
	return tokens
}
