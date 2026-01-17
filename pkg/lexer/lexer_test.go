package lexer

import (
	"testing"

	"github.com/caokhang91/buddhist-go/pkg/token"
)

func TestNextToken(t *testing.T) {
	input := `let five = 5;
let ten = 10;

let add = fn(x, y) {
  x + y;
};

let result = add(five, ten);
!-*5;
5 < 10 > 5;

if (5 < 10) {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9;
"foobar"
"foo bar"
[1, 2];
{"foo": "bar"}
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestKeywords(t *testing.T) {
	tests := []struct {
		input    string
		expected token.TokenType
	}{
		{"let", token.LET},
		{"const", token.CONST},
		{"fn", token.FUNCTION},
		{"if", token.IF},
		{"else", token.ELSE},
		{"return", token.RETURN},
		{"true", token.TRUE},
		{"false", token.FALSE},
		{"null", token.NULL},
		{"spawn", token.SPAWN},
		{"channel", token.CHANNEL},
		{"class", token.CLASS},
		{"import", token.IMPORT},
		{"export", token.EXPORT},
		{"from", token.FROM},
		{"try", token.TRY},
		{"catch", token.CATCH},
		{"finally", token.FINALLY},
		{"throw", token.THROW},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()
		if tok.Type != tt.expected {
			t.Errorf("keyword %q: expected %q, got %q", tt.input, tt.expected, tok.Type)
		}
	}
}

func TestFloat(t *testing.T) {
	input := "3.14 2.5 0.1"
	l := New(input)

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.FLOAT, "3.14"},
		{token.FLOAT, "2.5"},
		{token.FLOAT, "0.1"},
		{token.EOF, ""},
	}

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - type wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestOperators(t *testing.T) {
	input := "+-*/%<>!= ==!="
	l := New(input)

	tests := []struct {
		expectedType token.TokenType
	}{
		{token.PLUS},
		{token.MINUS},
		{token.ASTERISK},
		{token.SLASH},
		{token.MODULO},
		{token.LT},
		{token.GT},
		{token.NOT_EQ}, // != is parsed as NOT_EQ
		{token.EQ},     // == is parsed as EQ
		{token.NOT_EQ}, // != is parsed as NOT_EQ
	}

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - type wrong. expected=%q, got=%q", i, tt.expectedType, tok.Type)
		}
	}
}
