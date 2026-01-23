package parser

import (
	"fmt"
	"strconv"

	"github.com/caokhang91/buddhist-go/pkg/ast"
	"github.com/caokhang91/buddhist-go/pkg/lexer"
	"github.com/caokhang91/buddhist-go/pkg/token"
)

// Operator precedence levels
const (
	_ int = iota
	LOWEST
	ASSIGN      // =
	OR          // ||
	AND         // &&
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

var precedences = map[token.TokenType]int{
	token.ASSIGN:   ASSIGN,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LT_EQ:    LESSGREATER,
	token.GT_EQ:    LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.MODULO:   PRODUCT,
	token.AND:      AND,
	token.OR:       OR,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
	token.DOT:      CALL, // Property/method access has same precedence as function calls
	token.SEND:     ASSIGN, // Channel send should have same precedence as assignment
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser represents the parser
type Parser struct {
	l      lexer.TokenLexer // Accept interface for both standard and optimized lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// New creates a new Parser
func New(l lexer.TokenLexer) *Parser {
	p := &Parser{
		l:      l,
		errors: make([]string, 0, 4), // Pre-allocate error slice
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.BLOB, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.NULL, p.parseNullLiteral)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)
	p.registerPrefix(token.SPAWN, p.parseSpawnExpression)
	p.registerPrefix(token.CHANNEL, p.parseChannelExpression)
	p.registerPrefix(token.SEND, p.parseReceiveExpression)
	p.registerPrefix(token.LT, p.parseReceiveExpression) // Support < as prefix for channel receive
	p.registerPrefix(token.THIS, p.parseThisExpression)
	p.registerPrefix(token.SUPER, p.parseSuperExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LT_EQ, p.parseInfixExpression)
	p.registerInfix(token.GT_EQ, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseDotExpression) // Property/method access
	p.registerInfix(token.ASSIGN, p.parseAssignmentExpression)
	p.registerInfix(token.SEND, p.parseSendExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	// Fail-fast: return false immediately on error
	return false
}

// expectPeekFailFast is like expectPeek but stops parsing on error
func (p *Parser) expectPeekFailFast(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	// Fail-fast: stop parsing immediately
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("line %d: expected next token to be %s, got %s instead",
		p.peekToken.Line, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("line %d: no prefix parse function for %s found",
		p.curToken.Line, t)
	p.errors = append(p.errors, msg)
}

// Errors returns the parser errors
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
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

// ParseProgram parses the program and returns the AST
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	// Pre-allocate statements slice with reasonable capacity
	program.Statements = make([]ast.Statement, 0, 32)

	for !p.curTokenIs(token.EOF) {
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
	case token.PLACE:
		return p.parseLetStatement()
	case token.SET:
		return p.parseSetStatement()
	case token.CONST:
		return p.parseConstStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	case token.CLASS:
		return p.parseClassStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseSetStatement() *ast.SetStatement {
	stmt := &ast.SetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseConstStatement() *ast.ConstStatement {
	stmt := &ast.ConstStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	if !p.curTokenIs(token.SEMICOLON) {
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()

	// Parse init statement
	if !p.curTokenIs(token.SEMICOLON) {
		stmt.Init = p.parseStatement()
	}

	if !p.curTokenIs(token.SEMICOLON) {
		if !p.expectPeek(token.SEMICOLON) {
			return nil
		}
	}
	p.nextToken()

	// Parse condition
	if !p.curTokenIs(token.SEMICOLON) {
		stmt.Condition = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}
	p.nextToken()

	// Parse post statement
	if !p.curTokenIs(token.RPAREN) {
		stmt.Post = p.parseExpressionStatement()
	}

	if !p.curTokenIs(token.RPAREN) {
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}

	// Fail-fast: stop immediately on parsing errors
	if !p.expectPeekFailFast(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if stmt.Condition == nil {
		return nil // Fail-fast: stop if condition parsing failed
	}

	if !p.expectPeekFailFast(token.RPAREN) {
		return nil
	}

	if !p.expectPeekFailFast(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	if stmt.Body == nil {
		return nil // Fail-fast: stop if body parsing failed
	}

	// Check for optional until clause
	if p.peekTokenIs(token.UNTIL) {
		p.nextToken() // consume 'until'
		if !p.expectPeekFailFast(token.LPAREN) {
			return nil // Fail-fast: stop parsing on error
		}
		p.nextToken()
		stmt.Until = p.parseExpression(LOWEST)
		if stmt.Until == nil {
			return nil // Fail-fast: stop if expression parsing failed
		}
		if !p.expectPeekFailFast(token.RPAREN) {
			return nil // Fail-fast: stop parsing on error
		}
	}

	return stmt
}

func (p *Parser) parseBreakStatement() *ast.BreakStatement {
	stmt := &ast.BreakStatement{Token: p.curToken}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseContinueStatement() *ast.ContinueStatement {
	stmt := &ast.ContinueStatement{Token: p.curToken}
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) {
		peekPrec := p.peekPrecedence()
		
		// Special handling for ASSIGN: always allow it to parse (right-associative)
		// This allows property assignments like p.name = value to work correctly,
		// even when the left side has higher precedence (like DOT/CALL)
		if p.peekTokenIs(token.ASSIGN) {
			// Force ASSIGN to be parsed by making peekPrec higher than current precedence
			// This handles cases like p.name = value where p.name has CALL precedence
			peekPrec = precedence + 1
		}
		
		if precedence >= peekPrec {
			break
		}

		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("line %d: could not parse %q as integer",
			p.curToken.Line, p.curToken.Literal)
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
		msg := fmt.Sprintf("line %d: could not parse %q as float",
			p.curToken.Line, p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
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

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	// Support "if not (condition) then { ... }" syntax
	hasNot := p.peekTokenIs(token.NOT)
	var notToken token.Token
	if hasNot {
		p.nextToken() // consume "not"
		notToken = p.curToken // store the NOT token for later use
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	condition := p.parseExpression(LOWEST)

	// If "not" keyword was used, wrap the condition in a prefix expression with BANG operator
	if hasNot {
		expression.Condition = &ast.PrefixExpression{
			Token:    token.Token{Type: token.BANG, Literal: "!", Line: notToken.Line, Column: notToken.Column},
			Operator: "!",
			Right:    condition,
		}
	} else {
		expression.Condition = condition
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// Support optional "then" keyword (e.g., "if (condition) then { ... }")
	// If next token is THEN, consume it (optional)
	if p.peekTokenIs(token.THEN) {
		p.nextToken()
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		// Support "else if" - check if next token is IF
		if p.peekTokenIs(token.IF) {
			// Advance to IF token before parsing
			p.nextToken()
			// Parse "else if" as a nested if expression wrapped in a block statement
			ifExpr := p.parseIfExpression()
			if ifExpr == nil {
				return nil
			}
			// Wrap the if expression in a block statement containing an expression statement
			block := &ast.BlockStatement{
				Token:      token.Token{Type: token.LBRACE, Literal: "{"},
				Statements: []ast.Statement{
					&ast.ExpressionStatement{
						Token:      token.Token{Type: token.IF, Literal: "if"},
						Expression: ifExpr,
					},
				},
			}
			expression.Alternative = block
		} else {
			// Regular "else" block
			if !p.expectPeek(token.LBRACE) {
				return nil
			}
			expression.Alternative = p.parseBlockStatement()
		}
	}

	return expression
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	// Pre-allocate statements with reasonable capacity
	block.Statements = make([]ast.Statement, 0, 8)

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	// Check for optional function name
	if p.peekTokenIs(token.IDENT) {
		p.nextToken()
		lit.Name = p.curToken.Literal
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	// Pre-allocate with reasonable capacity for function parameters
	identifiers := make([]*ast.Identifier, 0, 4)

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	// Pre-allocate with reasonable capacity for function arguments
	list := make([]ast.Expression, 0, 4)

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = []ast.ArrayElement{}

	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		return array
	}

	p.nextToken()
	for !p.curTokenIs(token.RBRACKET) && !p.curTokenIs(token.EOF) {
		keyOrValue := p.parseExpression(LOWEST)
		if keyOrValue == nil {
			return nil
		}

		if p.peekTokenIs(token.ARROW) {
			p.nextToken()
			p.nextToken()
			value := p.parseExpression(LOWEST)
			if value == nil {
				return nil
			}
			array.Elements = append(array.Elements, ast.ArrayElement{Key: keyOrValue, Value: value})
		} else {
			array.Elements = append(array.Elements, ast.ArrayElement{Value: keyOrValue})
		}

		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			continue
		}

		if p.peekTokenIs(token.RBRACKET) {
			p.nextToken()
			break
		}

		if !p.expectPeek(token.RBRACKET) {
			return nil
		}
		break
	}
	return array
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	if p.peekTokenIs(token.RBRACKET) {
		p.nextToken()
		exp.Index = nil
		return exp
	}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)

		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return hash
}

func (p *Parser) parseSpawnExpression() ast.Expression {
	exp := &ast.SpawnExpression{Token: p.curToken}

	p.nextToken()
	exp.Function = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseChannelExpression() ast.Expression {
	exp := &ast.ChannelExpression{Token: p.curToken}

	// Check for buffered channel syntax: channel(bufferSize)
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken() // consume '('
		p.nextToken() // move to buffer size expression
		exp.BufferSize = p.parseExpression(LOWEST)
		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	}

	return exp
}

func (p *Parser) parseSendExpression(left ast.Expression) ast.Expression {
	exp := &ast.SendExpression{
		Token:   p.curToken,
		Channel: left,
	}

	p.nextToken()
	exp.Value = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseReceiveExpression() ast.Expression {
	exp := &ast.ReceiveExpression{Token: p.curToken}

	p.nextToken()
	exp.Channel = p.parseExpression(LOWEST)

	return exp
}

func (p *Parser) parseAssignmentExpression(left ast.Expression) ast.Expression {
	ident, ok := left.(*ast.Identifier)
	if ok {
		exp := &ast.AssignmentExpression{
			Token: p.curToken,
			Name:  ident,
		}

		p.nextToken()
		exp.Value = p.parseExpression(LOWEST)

		return exp
	}

	indexExp, ok := left.(*ast.IndexExpression)
	if ok {
		exp := &ast.IndexAssignmentExpression{
			Token:      p.curToken,
			IndexToken: indexExp.Token,
			Left:       indexExp.Left,
			Index:      indexExp.Index,
		}

		p.nextToken()
		exp.Value = p.parseExpression(LOWEST)

		return exp
	}

	msg := fmt.Sprintf("line %d: cannot assign to %s", p.curToken.Line, left.String())
	p.errors = append(p.errors, msg)
	return nil
}

// parseClassStatement parses a class declaration
func (p *Parser) parseClassStatement() *ast.ClassStatement {
	stmt := &ast.ClassStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// Check for 'extends' keyword
	if p.peekTokenIs(token.EXTENDS) {
		p.nextToken() // consume 'extends'
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		stmt.Parent = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseThisExpression parses the 'this' keyword
func (p *Parser) parseThisExpression() ast.Expression {
	return &ast.ThisExpression{Token: p.curToken}
}

// parseSuperExpression parses the 'super' keyword
func (p *Parser) parseSuperExpression() ast.Expression {
	return &ast.SuperExpression{Token: p.curToken}
}

// parseDotExpression parses property/method access: obj.property or obj.method()
func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{
		Token: p.curToken,
		Left:  left,
	}

	p.nextToken()

	if !p.curTokenIs(token.IDENT) {
		p.errors = append(p.errors, fmt.Sprintf("line %d: expected property name after '.'", p.curToken.Line))
		return nil
	}

	// Use the identifier as the index (property name)
	exp.Index = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	return exp
}
