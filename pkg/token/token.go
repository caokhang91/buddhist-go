package token

// TokenType represents the type of token
type TokenType string

// Token represents a lexical token
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// Token types
const (
	// Special tokens
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"

	// Identifiers + literals
	IDENT  TokenType = "IDENT"  // add, foobar, x, y, ...
	INT    TokenType = "INT"    // 1234567890
	FLOAT  TokenType = "FLOAT"  // 123.456
	STRING TokenType = "STRING" // "hello world"

	// Operators
	ASSIGN   TokenType = "="
	PLUS     TokenType = "+"
	MINUS    TokenType = "-"
	BANG     TokenType = "!"
	ASTERISK TokenType = "*"
	SLASH    TokenType = "/"
	MODULO   TokenType = "%"

	LT     TokenType = "<"
	GT     TokenType = ">"
	EQ     TokenType = "=="
	NOT_EQ TokenType = "!="
	LT_EQ  TokenType = "<="
	GT_EQ  TokenType = ">="

	// Logical operators
	AND TokenType = "&&"
	OR  TokenType = "||"

	// Delimiters
	COMMA     TokenType = ","
	SEMICOLON TokenType = ";"
	COLON     TokenType = ":"

	LPAREN   TokenType = "("
	RPAREN   TokenType = ")"
	LBRACE   TokenType = "{"
	RBRACE   TokenType = "}"
	LBRACKET TokenType = "["
	RBRACKET TokenType = "]"

	// Keywords
	FUNCTION TokenType = "FUNCTION"
	LET      TokenType = "LET"
	CONST    TokenType = "CONST"
	TRUE     TokenType = "TRUE"
	FALSE    TokenType = "FALSE"
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	RETURN   TokenType = "RETURN"
	FOR      TokenType = "FOR"
	WHILE    TokenType = "WHILE"
	BREAK    TokenType = "BREAK"
	CONTINUE TokenType = "CONTINUE"
	NULL     TokenType = "NULL"

	// Concurrency keywords (tận dụng Go Runtime)
	SPAWN   TokenType = "SPAWN"   // Tạo goroutine
	CHANNEL TokenType = "CHANNEL" // Khai báo channel
	SEND    TokenType = "<-"      // Gửi vào channel
	RECEIVE TokenType = "->"      // Nhận từ channel
)

// keywords maps keyword strings to their token types
var keywords = map[string]TokenType{
	"fn":       FUNCTION,
	"let":      LET,
	"const":    CONST,
	"true":     TRUE,
	"false":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"return":   RETURN,
	"for":      FOR,
	"while":    WHILE,
	"break":    BREAK,
	"continue": CONTINUE,
	"null":     NULL,
	"spawn":    SPAWN,
	"channel":  CHANNEL,
}

// LookupIdent checks if the identifier is a keyword
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// New creates a new token
func New(tokenType TokenType, literal string, line, column int) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    line,
		Column:  column,
	}
}
