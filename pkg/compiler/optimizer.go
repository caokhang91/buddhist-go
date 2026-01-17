package compiler

import (
	"github.com/caokhang91/buddhist-go/pkg/ast"
	"github.com/caokhang91/buddhist-go/pkg/object"
)

// EnableOptimizations controls whether optimizations are applied
var EnableOptimizations = true

// Optimizer performs compile-time optimizations on AST
type Optimizer struct{}

// NewOptimizer creates a new optimizer
func NewOptimizer() *Optimizer {
	return &Optimizer{}
}

// OptimizeProgram is a package-level function for optimizing programs
// It applies constant folding and other optimizations when enabled
func OptimizeProgram(program *ast.Program) *ast.Program {
	if !EnableOptimizations {
		return program
	}
	
	o := &Optimizer{}
	optimizedStatements := make([]ast.Statement, len(program.Statements))
	for i, stmt := range program.Statements {
		optimizedStatements[i] = o.optimizeStatement(stmt)
	}
	return &ast.Program{Statements: optimizedStatements}
}

// OptimizeProgramMethod optimizes the entire program (method version)
func (o *Optimizer) OptimizeProgramMethod(program *ast.Program) *ast.Program {
	optimizedStatements := make([]ast.Statement, len(program.Statements))
	for i, stmt := range program.Statements {
		optimizedStatements[i] = o.optimizeStatement(stmt)
	}
	return &ast.Program{Statements: optimizedStatements}
}

func (o *Optimizer) optimizeStatement(stmt ast.Statement) ast.Statement {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return &ast.ExpressionStatement{
			Token:      s.Token,
			Expression: o.optimizeExpression(s.Expression),
		}
	case *ast.LetStatement:
		return &ast.LetStatement{
			Token: s.Token,
			Name:  s.Name,
			Value: o.optimizeExpression(s.Value),
		}
	case *ast.ConstStatement:
		return &ast.ConstStatement{
			Token: s.Token,
			Name:  s.Name,
			Value: o.optimizeExpression(s.Value),
		}
	case *ast.ReturnStatement:
		if s.ReturnValue != nil {
			return &ast.ReturnStatement{
				Token:       s.Token,
				ReturnValue: o.optimizeExpression(s.ReturnValue),
			}
		}
		return s
	case *ast.BlockStatement:
		return o.optimizeBlockStatement(s)
	case *ast.WhileStatement:
		return &ast.WhileStatement{
			Token:     s.Token,
			Condition: o.optimizeExpression(s.Condition),
			Body:      o.optimizeBlockStatement(s.Body),
		}
	case *ast.ForStatement:
		var init ast.Statement
		var post ast.Statement
		var condition ast.Expression
		if s.Init != nil {
			init = o.optimizeStatement(s.Init)
		}
		if s.Condition != nil {
			condition = o.optimizeExpression(s.Condition)
		}
		if s.Post != nil {
			post = o.optimizeStatement(s.Post)
		}
		return &ast.ForStatement{
			Token:     s.Token,
			Init:      init,
			Condition: condition,
			Post:      post,
			Body:      o.optimizeBlockStatement(s.Body),
		}
	default:
		return stmt
	}
}

func (o *Optimizer) optimizeBlockStatement(block *ast.BlockStatement) *ast.BlockStatement {
	if block == nil {
		return nil
	}
	optimizedStatements := make([]ast.Statement, len(block.Statements))
	for i, stmt := range block.Statements {
		optimizedStatements[i] = o.optimizeStatement(stmt)
	}
	return &ast.BlockStatement{
		Token:      block.Token,
		Statements: optimizedStatements,
	}
}

func (o *Optimizer) optimizeExpression(expr ast.Expression) ast.Expression {
	if expr == nil {
		return nil
	}
	
	switch e := expr.(type) {
	case *ast.InfixExpression:
		return o.optimizeInfixExpression(e)
	case *ast.PrefixExpression:
		return o.optimizePrefixExpression(e)
	case *ast.IfExpression:
		return o.optimizeIfExpression(e)
	case *ast.FunctionLiteral:
		return &ast.FunctionLiteral{
			Token:      e.Token,
			Name:       e.Name,
			Parameters: e.Parameters,
			Body:       o.optimizeBlockStatement(e.Body),
		}
	case *ast.CallExpression:
		args := make([]ast.Expression, len(e.Arguments))
		for i, arg := range e.Arguments {
			args[i] = o.optimizeExpression(arg)
		}
		return &ast.CallExpression{
			Token:     e.Token,
			Function:  o.optimizeExpression(e.Function),
			Arguments: args,
		}
	case *ast.ArrayLiteral:
		elements := make([]ast.ArrayElement, len(e.Elements))
		for i, el := range e.Elements {
			elements[i] = ast.ArrayElement{
				Key:   o.optimizeExpression(el.Key),
				Value: o.optimizeExpression(el.Value),
			}
		}
		return &ast.ArrayLiteral{
			Token:    e.Token,
			Elements: elements,
		}
	case *ast.HashLiteral:
		pairs := make(map[ast.Expression]ast.Expression)
		for k, v := range e.Pairs {
			pairs[o.optimizeExpression(k)] = o.optimizeExpression(v)
		}
		return &ast.HashLiteral{
			Token: e.Token,
			Pairs: pairs,
		}
	case *ast.IndexExpression:
		return &ast.IndexExpression{
			Token: e.Token,
			Left:  o.optimizeExpression(e.Left),
			Index: o.optimizeExpression(e.Index),
		}
	case *ast.AssignmentExpression:
		return &ast.AssignmentExpression{
			Token: e.Token,
			Name:  e.Name,
			Value: o.optimizeExpression(e.Value),
		}
	default:
		return expr
	}
}

// optimizeInfixExpression performs constant folding on infix expressions
func (o *Optimizer) optimizeInfixExpression(expr *ast.InfixExpression) ast.Expression {
	left := o.optimizeExpression(expr.Left)
	right := o.optimizeExpression(expr.Right)

	// Try constant folding for integer operations
	leftInt, leftIsInt := left.(*ast.IntegerLiteral)
	rightInt, rightIsInt := right.(*ast.IntegerLiteral)

	if leftIsInt && rightIsInt {
		result := o.foldIntegerOperation(expr.Operator, leftInt.Value, rightInt.Value)
		if result != nil {
			return result
		}
	}

	// Try constant folding for float operations
	leftFloat, rightFloat, bothFloat := o.getFloatValues(left, right)
	if bothFloat {
		result := o.foldFloatOperation(expr.Operator, leftFloat, rightFloat)
		if result != nil {
			return result
		}
	}

	// Try constant folding for string concatenation
	leftStr, leftIsStr := left.(*ast.StringLiteral)
	rightStr, rightIsStr := right.(*ast.StringLiteral)

	if leftIsStr && rightIsStr && expr.Operator == "+" {
		return &ast.StringLiteral{
			Token: expr.Token,
			Value: leftStr.Value + rightStr.Value,
		}
	}

	// Try constant folding for boolean operations
	leftBool, leftIsBool := left.(*ast.Boolean)
	rightBool, rightIsBool := right.(*ast.Boolean)

	if leftIsBool && rightIsBool {
		result := o.foldBooleanOperation(expr.Operator, leftBool.Value, rightBool.Value)
		if result != nil {
			return result
		}
	}

	// Return optimized expression if folding didn't apply
	return &ast.InfixExpression{
		Token:    expr.Token,
		Left:     left,
		Operator: expr.Operator,
		Right:    right,
	}
}

func (o *Optimizer) foldIntegerOperation(op string, left, right int64) ast.Expression {
	var result int64
	var boolResult bool
	isBoolOp := false

	switch op {
	case "+":
		result = left + right
	case "-":
		result = left - right
	case "*":
		result = left * right
	case "/":
		if right == 0 {
			return nil // Can't fold division by zero
		}
		result = left / right
	case "%":
		if right == 0 {
			return nil // Can't fold modulo by zero
		}
		result = left % right
	case "<":
		boolResult = left < right
		isBoolOp = true
	case ">":
		boolResult = left > right
		isBoolOp = true
	case "<=":
		boolResult = left <= right
		isBoolOp = true
	case ">=":
		boolResult = left >= right
		isBoolOp = true
	case "==":
		boolResult = left == right
		isBoolOp = true
	case "!=":
		boolResult = left != right
		isBoolOp = true
	default:
		return nil
	}

	if isBoolOp {
		return &ast.Boolean{Value: boolResult}
	}
	return &ast.IntegerLiteral{Value: result}
}

func (o *Optimizer) getFloatValues(left, right ast.Expression) (float64, float64, bool) {
	var leftVal, rightVal float64
	var hasLeft, hasRight bool

	switch l := left.(type) {
	case *ast.FloatLiteral:
		leftVal = l.Value
		hasLeft = true
	case *ast.IntegerLiteral:
		leftVal = float64(l.Value)
		hasLeft = true
	}

	switch r := right.(type) {
	case *ast.FloatLiteral:
		rightVal = r.Value
		hasRight = true
	case *ast.IntegerLiteral:
		rightVal = float64(r.Value)
		hasRight = true
	}

	// Only return true if at least one is a float literal
	_, leftIsFloat := left.(*ast.FloatLiteral)
	_, rightIsFloat := right.(*ast.FloatLiteral)

	return leftVal, rightVal, hasLeft && hasRight && (leftIsFloat || rightIsFloat)
}

func (o *Optimizer) foldFloatOperation(op string, left, right float64) ast.Expression {
	var result float64
	var boolResult bool
	isBoolOp := false

	switch op {
	case "+":
		result = left + right
	case "-":
		result = left - right
	case "*":
		result = left * right
	case "/":
		if right == 0 {
			return nil
		}
		result = left / right
	case "<":
		boolResult = left < right
		isBoolOp = true
	case ">":
		boolResult = left > right
		isBoolOp = true
	case "<=":
		boolResult = left <= right
		isBoolOp = true
	case ">=":
		boolResult = left >= right
		isBoolOp = true
	case "==":
		boolResult = left == right
		isBoolOp = true
	case "!=":
		boolResult = left != right
		isBoolOp = true
	default:
		return nil
	}

	if isBoolOp {
		return &ast.Boolean{Value: boolResult}
	}
	return &ast.FloatLiteral{Value: result}
}

func (o *Optimizer) foldBooleanOperation(op string, left, right bool) ast.Expression {
	switch op {
	case "&&":
		return &ast.Boolean{Value: left && right}
	case "||":
		return &ast.Boolean{Value: left || right}
	case "==":
		return &ast.Boolean{Value: left == right}
	case "!=":
		return &ast.Boolean{Value: left != right}
	default:
		return nil
	}
}

func (o *Optimizer) optimizePrefixExpression(expr *ast.PrefixExpression) ast.Expression {
	right := o.optimizeExpression(expr.Right)

	switch expr.Operator {
	case "-":
		// Fold negative integer
		if intLit, ok := right.(*ast.IntegerLiteral); ok {
			return &ast.IntegerLiteral{Value: -intLit.Value}
		}
		// Fold negative float
		if floatLit, ok := right.(*ast.FloatLiteral); ok {
			return &ast.FloatLiteral{Value: -floatLit.Value}
		}
	case "!":
		// Fold boolean negation
		if boolLit, ok := right.(*ast.Boolean); ok {
			return &ast.Boolean{Value: !boolLit.Value}
		}
	}

	return &ast.PrefixExpression{
		Token:    expr.Token,
		Operator: expr.Operator,
		Right:    right,
	}
}

func (o *Optimizer) optimizeIfExpression(expr *ast.IfExpression) ast.Expression {
	condition := o.optimizeExpression(expr.Condition)

	// If condition is a constant boolean, eliminate dead branch
	if boolLit, ok := condition.(*ast.Boolean); ok {
		if boolLit.Value {
			// Condition is always true, return consequence
			// This is a dead code elimination optimization
			// For now, we'll keep the if expression but optimize branches
		} else {
			// Condition is always false
			// For now, we'll keep the if expression but optimize branches
		}
	}

	return &ast.IfExpression{
		Token:       expr.Token,
		Condition:   condition,
		Consequence: o.optimizeBlockStatement(expr.Consequence),
		Alternative: o.optimizeBlockStatement(expr.Alternative),
	}
}

// FoldConstants is a helper function to perform constant folding
// This can be called during compilation to evaluate constant expressions
func FoldConstants(left, right object.Object, op string) object.Object {
	leftInt, leftIsInt := left.(*object.Integer)
	rightInt, rightIsInt := right.(*object.Integer)

	if leftIsInt && rightIsInt {
		switch op {
		case "+":
			return object.GetCachedInteger(leftInt.Value + rightInt.Value)
		case "-":
			return object.GetCachedInteger(leftInt.Value - rightInt.Value)
		case "*":
			return object.GetCachedInteger(leftInt.Value * rightInt.Value)
		case "/":
			if rightInt.Value != 0 {
				return object.GetCachedInteger(leftInt.Value / rightInt.Value)
			}
		case "%":
			if rightInt.Value != 0 {
				return object.GetCachedInteger(leftInt.Value % rightInt.Value)
			}
		}
	}

	return nil
}
