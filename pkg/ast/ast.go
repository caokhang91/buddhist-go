package ast

import (
	"bytes"
	"strings"

	"github.com/caokhang91/buddhist-go/pkg/token"
)

// Node represents a node in the AST
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

// Program is the root node of every AST
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
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// LetStatement represents a let statement: let x = 5;
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	out.WriteString(" = ")
	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// ConstStatement represents a const statement: const x = 5;
type ConstStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (cs *ConstStatement) statementNode()       {}
func (cs *ConstStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ConstStatement) String() string {
	var out bytes.Buffer
	out.WriteString(cs.TokenLiteral() + " ")
	out.WriteString(cs.Name.String())
	out.WriteString(" = ")
	if cs.Value != nil {
		out.WriteString(cs.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// ReturnStatement represents a return statement: return 5;
type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer
	out.WriteString(rs.TokenLiteral() + " ")
	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}
	out.WriteString(";")
	return out.String()
}

// ExpressionStatement represents an expression statement
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// BlockStatement represents a block of statements
type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// Identifier represents an identifier
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral represents an integer literal
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral represents a float literal
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral represents a string literal
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

// Boolean represents a boolean literal
type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }

// NullLiteral represents a null literal
type NullLiteral struct {
	Token token.Token
}

func (nl *NullLiteral) expressionNode()      {}
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullLiteral) String() string       { return "null" }

// PrefixExpression represents a prefix expression: !x, -5
type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. !
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression represents an infix expression: 5 + 5
type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// IfExpression represents an if expression
type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}

// ForStatement represents a for loop
type ForStatement struct {
	Token     token.Token
	Init      Statement
	Condition Expression
	Post      Statement
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for (")
	if fs.Init != nil {
		out.WriteString(fs.Init.String())
	}
	out.WriteString("; ")
	if fs.Condition != nil {
		out.WriteString(fs.Condition.String())
	}
	out.WriteString("; ")
	if fs.Post != nil {
		out.WriteString(fs.Post.String())
	}
	out.WriteString(") ")
	out.WriteString(fs.Body.String())
	return out.String()
}

// WhileStatement represents a while loop
type WhileStatement struct {
	Token     token.Token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while (")
	out.WriteString(ws.Condition.String())
	out.WriteString(") ")
	out.WriteString(ws.Body.String())
	return out.String()
}

// BreakStatement represents a break statement
type BreakStatement struct {
	Token token.Token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return "break;" }

// ContinueStatement represents a continue statement
type ContinueStatement struct {
	Token token.Token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return "continue;" }

// FunctionLiteral represents a function literal
type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Name       string      // Optional function name
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(fl.TokenLiteral())
	if fl.Name != "" {
		out.WriteString(" " + fl.Name)
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fl.Body.String())
	return out.String()
}

// CallExpression represents a function call
type CallExpression struct {
	Token     token.Token // The '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// ArrayLiteral represents an array literal: [1, 2, 3]
type ArrayLiteral struct {
	Token    token.Token // the '[' token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer
	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// IndexExpression represents an index expression: arr[0]
type IndexExpression struct {
	Token token.Token // The [ token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}

// HashLiteral represents a hash/map literal: {"key": value}
type HashLiteral struct {
	Token token.Token // the '{' token
	Pairs map[Expression]Expression
}

func (hl *HashLiteral) expressionNode()      {}
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }
func (hl *HashLiteral) String() string {
	var out bytes.Buffer
	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// SpawnExpression represents a spawn expression for concurrency
type SpawnExpression struct {
	Token    token.Token // the 'spawn' token
	Function Expression  // The function to execute concurrently
}

func (se *SpawnExpression) expressionNode()      {}
func (se *SpawnExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SpawnExpression) String() string {
	var out bytes.Buffer
	out.WriteString("spawn ")
	out.WriteString(se.Function.String())
	return out.String()
}

// ChannelExpression represents a channel declaration
type ChannelExpression struct {
	Token token.Token // the 'channel' token
}

func (ce *ChannelExpression) expressionNode()      {}
func (ce *ChannelExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *ChannelExpression) String() string       { return "channel" }

// SendExpression represents sending a value to a channel: ch <- value
type SendExpression struct {
	Token   token.Token // the '<-' token
	Channel Expression
	Value   Expression
}

func (se *SendExpression) expressionNode()      {}
func (se *SendExpression) TokenLiteral() string { return se.Token.Literal }
func (se *SendExpression) String() string {
	var out bytes.Buffer
	out.WriteString(se.Channel.String())
	out.WriteString(" <- ")
	out.WriteString(se.Value.String())
	return out.String()
}

// ReceiveExpression represents receiving a value from a channel: <- ch
type ReceiveExpression struct {
	Token   token.Token // the '<-' token
	Channel Expression
}

func (re *ReceiveExpression) expressionNode()      {}
func (re *ReceiveExpression) TokenLiteral() string { return re.Token.Literal }
func (re *ReceiveExpression) String() string {
	var out bytes.Buffer
	out.WriteString("<- ")
	out.WriteString(re.Channel.String())
	return out.String()
}

// AssignmentExpression represents an assignment: x = 5
type AssignmentExpression struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (ae *AssignmentExpression) expressionNode()      {}
func (ae *AssignmentExpression) TokenLiteral() string { return ae.Token.Literal }
func (ae *AssignmentExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ae.Name.String())
	out.WriteString(" = ")
	out.WriteString(ae.Value.String())
	return out.String()
}

// ClassStatement represents a class declaration
type ClassStatement struct {
	Token token.Token // the 'class' token
	Name  *Identifier
	Body  *BlockStatement // Contains method definitions
}

func (cs *ClassStatement) statementNode()       {}
func (cs *ClassStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ClassStatement) String() string {
	var out bytes.Buffer
	out.WriteString("class ")
	out.WriteString(cs.Name.String())
	out.WriteString(" ")
	out.WriteString(cs.Body.String())
	return out.String()
}

// ThisExpression represents 'this' keyword
type ThisExpression struct {
	Token token.Token // the 'this' token
}

func (te *ThisExpression) expressionNode()      {}
func (te *ThisExpression) TokenLiteral() string { return te.Token.Literal }
func (te *ThisExpression) String() string       { return "this" }

// ImportStatement represents an import statement: import { name } from "module"
type ImportStatement struct {
	Token    token.Token // the 'import' token
	Names    []*Identifier
	FromPath string // The module path/name
}

func (is *ImportStatement) statementNode()       {}
func (is *ImportStatement) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStatement) String() string {
	var out bytes.Buffer
	out.WriteString("import { ")
	for i, name := range is.Names {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(name.String())
	}
	out.WriteString(" } from \"")
	out.WriteString(is.FromPath)
	out.WriteString("\"")
	return out.String()
}

// ExportStatement represents an export statement: export let x = 5;
type ExportStatement struct {
	Token     token.Token // the 'export' token
	Statement Statement   // The statement being exported
}

func (es *ExportStatement) statementNode()       {}
func (es *ExportStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExportStatement) String() string {
	var out bytes.Buffer
	out.WriteString("export ")
	out.WriteString(es.Statement.String())
	return out.String()
}

// TryExpression represents a try/catch/finally block
type TryExpression struct {
	Token        token.Token // the 'try' token
	TryBlock     *BlockStatement
	CatchVar     *Identifier // Variable name for caught error (can be nil)
	CatchBlock   *BlockStatement
	FinallyBlock *BlockStatement // Can be nil
}

func (te *TryExpression) expressionNode()      {}
func (te *TryExpression) TokenLiteral() string { return te.Token.Literal }
func (te *TryExpression) String() string {
	var out bytes.Buffer
	out.WriteString("try ")
	out.WriteString(te.TryBlock.String())
	if te.CatchBlock != nil {
		out.WriteString(" catch")
		if te.CatchVar != nil {
			out.WriteString(" (")
			out.WriteString(te.CatchVar.String())
			out.WriteString(")")
		}
		out.WriteString(" ")
		out.WriteString(te.CatchBlock.String())
	}
	if te.FinallyBlock != nil {
		out.WriteString(" finally ")
		out.WriteString(te.FinallyBlock.String())
	}
	return out.String()
}

// ThrowStatement represents a throw statement: throw error;
type ThrowStatement struct {
	Token token.Token // the 'throw' token
	Value Expression  // The value/error to throw
}

func (ts *ThrowStatement) statementNode()       {}
func (ts *ThrowStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *ThrowStatement) String() string {
	var out bytes.Buffer
	out.WriteString("throw ")
	if ts.Value != nil {
		out.WriteString(ts.Value.String())
	}
	return out.String()
}
