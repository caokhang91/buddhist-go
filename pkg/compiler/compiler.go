package compiler

import (
	"fmt"
	"sort"

	"github.com/caokhang91/buddhist-go/pkg/ast"
	"github.com/caokhang91/buddhist-go/pkg/code"
	"github.com/caokhang91/buddhist-go/pkg/object"
)

// Bytecode represents compiled bytecode
type Bytecode struct {
	Instructions code.Instructions
	Constants    []object.Object
}

// EmittedInstruction represents an emitted instruction
type EmittedInstruction struct {
	Opcode   code.Opcode
	Position int
}

// CompilationScope represents a compilation scope
type CompilationScope struct {
	instructions        code.Instructions
	lastInstruction     EmittedInstruction
	previousInstruction EmittedInstruction
}

// Compiler compiles AST to bytecode
type Compiler struct {
	constants   []object.Object
	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int
}

// New creates a new Compiler
func New() *Compiler {
	mainScope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()

	// Define builtin functions
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	return &Compiler{
		constants:   []object.Object{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
	}
}

// NewWithState creates a new compiler with existing state
func NewWithState(s *SymbolTable, constants []object.Object) *Compiler {
	compiler := New()
	compiler.symbolTable = s
	compiler.constants = constants
	return compiler
}

// Compile compiles an AST node
func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		// Apply optimizations before compilation (optional, can be disabled for debugging)
		// The optimizer performs constant folding and other compile-time optimizations
		optimizedProgram := OptimizeProgram(node)
		
		for _, s := range optimizedProgram.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(code.OpPop)

	case *ast.InfixExpression:
		// Handle < and <= by swapping operands
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThan)
			return nil
		}
		if node.Operator == "<=" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}
			err = c.Compile(node.Left)
			if err != nil {
				return err
			}
			c.emit(code.OpGreaterThanOrEqual)
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.emit(code.OpAdd)
		case "-":
			c.emit(code.OpSub)
		case "*":
			c.emit(code.OpMul)
		case "/":
			c.emit(code.OpDiv)
		case "%":
			c.emit(code.OpMod)
		case ">":
			c.emit(code.OpGreaterThan)
		case ">=":
			c.emit(code.OpGreaterThanOrEqual)
		case "==":
			c.emit(code.OpEqual)
		case "!=":
			c.emit(code.OpNotEqual)
		case "&&":
			c.emit(code.OpAnd)
		case "||":
			c.emit(code.OpOr)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}
		switch node.Operator {
		case "!":
			c.emit(code.OpBang)
		case "-":
			c.emit(code.OpMinus)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral:
		integer := &object.Integer{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(integer))

	case *ast.FloatLiteral:
		float := &object.Float{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(float))

	case *ast.StringLiteral:
		str := &object.String{Value: node.Value}
		c.emit(code.OpConstant, c.addConstant(str))

	case *ast.Boolean:
		if node.Value {
			c.emit(code.OpTrue)
		} else {
			c.emit(code.OpFalse)
		}

	case *ast.NullLiteral:
		c.emit(code.OpNull)

	case *ast.IfExpression:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit jump if not truthy with bogus value
		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.removeLastPop()
		}

		// Emit jump with bogus value
		jumpPos := c.emit(code.OpJump, 9999)

		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative == nil {
			c.emit(code.OpNull)
		} else {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}
			if c.lastInstructionIs(code.OpPop) {
				c.removeLastPop()
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.LetStatement:
		symbol := c.symbolTable.Define(node.Name.Value)
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
			c.emit(code.OpGetGlobal, symbol.Index) // Push the assigned value back
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
			c.emit(code.OpGetLocal, symbol.Index) // Push the assigned value back
		}

	case *ast.ConstStatement:
		symbol := c.symbolTable.Define(node.Name.Value)
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}

	case *ast.ClassStatement:
		// Compile class body to extract methods and properties
		methods := make(map[string]*object.CompiledFunction)
		properties := []string{}
		var constructor *object.CompiledFunction

		// Note: Parent class resolution will be done at runtime
		// because parent class might be defined after this class
		// We store the parent name in the class and resolve it when needed

		// Process class body statements
		for _, stmt := range node.Body.Statements {
			switch s := stmt.(type) {
			case *ast.LetStatement:
				// Property declaration: let name = value;
				// Just record the property name, don't compile the statement
				properties = append(properties, s.Name.Value)
			case *ast.ExpressionStatement:
				// Could be a function literal (method) or other expression
				if fnLit, ok := s.Expression.(*ast.FunctionLiteral); ok {
					// This is a method definition
					c.enterScope()
					// 'this' will be passed as first argument, so define it as first parameter
					c.symbolTable.Define("this")
					
					for _, p := range fnLit.Parameters {
						c.symbolTable.Define(p.Value)
					}

					err := c.Compile(fnLit.Body)
					if err != nil {
						return err
					}

					if c.lastInstructionIs(code.OpPop) {
						c.replaceLastPopWithReturn()
					}
					if !c.lastInstructionIs(code.OpReturnValue) {
						c.emit(code.OpReturn)
					}

					numLocals := c.symbolTable.numDefinitions
					instructions := c.leaveScope()

					compiledFn := &object.CompiledFunction{
						Instructions:  instructions,
						NumLocals:     numLocals,
						NumParameters: len(fnLit.Parameters) + 1, // +1 for 'this'
					}

					methodName := fnLit.Name
					if methodName == "" {
						// Anonymous method - skip
						continue
					}
					
					// Check if this is a constructor (named 'init' or 'constructor')
					if methodName == "init" || methodName == "constructor" {
						constructor = compiledFn
					} else {
						methods[methodName] = compiledFn
					}
				}
			}
		}

		// Create class object
		parentName := ""
		if node.Parent != nil {
			parentName = node.Parent.Value
		}
		class := &object.Class{
			Name:       node.Name.Value,
			Methods:    methods,
			Properties: properties,
			Parent:     nil, // Will be resolved at runtime
			ParentName: parentName,
		}
		
		// Store constructor separately if exists
		if constructor != nil {
			methods["init"] = constructor
		}

		// Add class to constants and emit OpClass
		classIndex := c.addConstant(class)
		c.emit(code.OpClass, classIndex)

		// Store class in global scope
		symbol := c.symbolTable.Define(node.Name.Value)
		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
		}
		
		// If parent exists, we need to resolve it after both classes are defined
		// For now, we'll resolve it at runtime when the class is instantiated

	case *ast.ThisExpression:
		// Push 'this' onto stack
		c.emit(code.OpThis)

	case *ast.SuperExpression:
		// Push 'super' onto stack
		c.emit(code.OpSuper)

	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}
		c.loadSymbol(symbol)

	case *ast.AssignmentExpression:
		symbol, ok := c.symbolTable.Resolve(node.Name.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Name.Value)
		}

		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		if symbol.Scope == GlobalScope {
			c.emit(code.OpSetGlobal, symbol.Index)
			c.emit(code.OpGetGlobal, symbol.Index) // Push the assigned value back
		} else {
			c.emit(code.OpSetLocal, symbol.Index)
			c.emit(code.OpGetLocal, symbol.Index) // Push the assigned value back
		}

	case *ast.ArrayLiteral:
		hasKeys := false
		for _, el := range node.Elements {
			if el.Key != nil {
				hasKeys = true
				break
			}
		}
		if hasKeys {
			for _, el := range node.Elements {
				if el.Key != nil {
					err := c.Compile(el.Key)
					if err != nil {
						return err
					}
				} else {
					c.emit(code.OpNull)
				}
				err := c.Compile(el.Value)
				if err != nil {
					return err
				}
			}
			c.emit(code.OpPHPArray, len(node.Elements))
		} else {
			for _, el := range node.Elements {
				err := c.Compile(el.Value)
				if err != nil {
					return err
				}
			}
			c.emit(code.OpArray, len(node.Elements))
		}

	case *ast.HashLiteral:
		keys := []ast.Expression{}
		for k := range node.Pairs {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, k := range keys {
			err := c.Compile(k)
			if err != nil {
				return err
			}
			err = c.Compile(node.Pairs[k])
			if err != nil {
				return err
			}
		}
		c.emit(code.OpHash, len(node.Pairs)*2)

	case *ast.IndexExpression:
		if node.Index == nil {
			return fmt.Errorf("index expression requires index")
		}
		
		// Check if this is property access (obj.property)
		if ident, ok := node.Index.(*ast.Identifier); ok {
			// Property access - use OpGetProperty
			err := c.Compile(node.Left)
			if err != nil {
				return err
			}
			// Push property name as string
			propName := &object.String{Value: ident.Value}
			propNameIndex := c.addConstant(propName)
			c.emit(code.OpConstant, propNameIndex)
			c.emit(code.OpGetProperty)
			return nil
		}
		
		// Regular index access (arr[index] or hash[key])
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		err = c.Compile(node.Index)
		if err != nil {
			return err
		}
		c.emit(code.OpIndex)

	case *ast.IndexAssignmentExpression:
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}
		if node.Index == nil {
			err = c.Compile(node.Value)
			if err != nil {
				return err
			}
			c.emit(code.OpArrayPush)
		} else {
			// Check if this is property assignment (obj.property = value)
			if ident, ok := node.Index.(*ast.Identifier); ok {
				// Property assignment - use OpSetProperty
				propName := &object.String{Value: ident.Value}
				propNameIndex := c.addConstant(propName)
				c.emit(code.OpConstant, propNameIndex)
				err = c.Compile(node.Value)
				if err != nil {
					return err
				}
				c.emit(code.OpSetProperty)
				// Push the assigned value back (it's already on stack from OpSetProperty)
				return nil
			}
			
			// Regular index assignment (arr[index] = value)
			err = c.Compile(node.Index)
			if err != nil {
				return err
			}
			err = c.Compile(node.Value)
			if err != nil {
				return err
			}
			c.emit(code.OpSetIndex)
		}

	case *ast.FunctionLiteral:
		c.enterScope()

		if node.Name != "" {
			c.symbolTable.DefineFunctionName(node.Name)
		}

		for _, p := range node.Parameters {
			c.symbolTable.Define(p.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		if c.lastInstructionIs(code.OpPop) {
			c.replaceLastPopWithReturn()
		}
		if !c.lastInstructionIs(code.OpReturnValue) {
			c.emit(code.OpReturn)
		}

		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numDefinitions
		instructions := c.leaveScope()

		for _, s := range freeSymbols {
			c.loadSymbol(s)
		}

		compiledFn := &object.CompiledFunction{
			Instructions:  instructions,
			NumLocals:     numLocals,
			NumParameters: len(node.Parameters),
		}
		fnIndex := c.addConstant(compiledFn)
		c.emit(code.OpClosure, fnIndex, len(freeSymbols))

	case *ast.ReturnStatement:
		if node.ReturnValue != nil {
			err := c.Compile(node.ReturnValue)
			if err != nil {
				return err
			}
		}
		c.emit(code.OpReturnValue)

	case *ast.SendStatement:
		err := c.Compile(node.Channel)
		if err != nil {
			return err
		}
		err = c.Compile(node.Value)
		if err != nil {
			return err
		}
		c.emit(code.OpSend)

	case *ast.CallExpression:
		// Check if this is a method call (obj.method())
		if indexExp, ok := node.Function.(*ast.IndexExpression); ok {
			if ident, ok := indexExp.Index.(*ast.Identifier); ok {
				// This is obj.method() - compile as method call
				err := c.Compile(indexExp.Left)
				if err != nil {
					return err
				}
				// Push method name as string
				methodName := &object.String{Value: ident.Value}
				methodNameIndex := c.addConstant(methodName)
				c.emit(code.OpConstant, methodNameIndex)

				for _, a := range node.Arguments {
					err := c.Compile(a)
					if err != nil {
						return err
					}
				}

				c.emit(code.OpCallMethod, len(node.Arguments))
				return nil
			}
		}

		// Regular function call
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, a := range node.Arguments {
			err := c.Compile(a)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpCall, len(node.Arguments))

	case *ast.WhileStatement:
		loopStart := len(c.currentInstructions())

		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		jumpNotTruthyPos := c.emit(code.OpJumpNotTruthy, 9999)

		err = c.Compile(node.Body)
		if err != nil {
			return err
		}

		c.emit(code.OpJump, loopStart)

		afterLoopPos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterLoopPos)

	case *ast.ForStatement:
		// Compile init
		if node.Init != nil {
			err := c.Compile(node.Init)
			if err != nil {
				return err
			}
		}

		loopStart := len(c.currentInstructions())

		// Compile condition
		var jumpNotTruthyPos int
		if node.Condition != nil {
			err := c.Compile(node.Condition)
			if err != nil {
				return err
			}
			jumpNotTruthyPos = c.emit(code.OpJumpNotTruthy, 9999)
		}

		// Compile body
		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		// Compile post
		if node.Post != nil {
			err := c.Compile(node.Post)
			if err != nil {
				return err
			}
		}

		c.emit(code.OpJump, loopStart)

		if node.Condition != nil {
			afterLoopPos := len(c.currentInstructions())
			c.changeOperand(jumpNotTruthyPos, afterLoopPos)
		}

	case *ast.BreakStatement:
		c.emit(code.OpBreak)

	case *ast.ContinueStatement:
		c.emit(code.OpContinue)

	case *ast.SpawnExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}
		c.emit(code.OpSpawn)

	case *ast.ChannelExpression:
		if node.BufferSize != nil {
			err := c.Compile(node.BufferSize)
			if err != nil {
				return err
			}
			c.emit(code.OpChannelBuffered)
		} else {
			c.emit(code.OpChannel)
		}

	case *ast.SendExpression:
		err := c.Compile(node.Channel)
		if err != nil {
			return err
		}
		err = c.Compile(node.Value)
		if err != nil {
			return err
		}
		c.emit(code.OpSend)

	case *ast.ReceiveExpression:
		err := c.Compile(node.Channel)
		if err != nil {
			return err
		}
		c.emit(code.OpReceive)
	}

	return nil
}

// Bytecode returns the compiled bytecode
func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) addConstant(obj object.Object) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

func (c *Compiler) emit(op code.Opcode, operands ...int) int {
	ins := code.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)
	c.scopes[c.scopeIndex].instructions = updatedInstructions
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}
	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op code.Opcode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()
	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := code.Opcode(c.currentInstructions()[opPos])
	newInstruction := code.Make(op, operand)
	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) currentInstructions() code.Instructions {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        code.Instructions{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++
	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() code.Instructions {
	instructions := c.currentInstructions()
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.symbolTable = c.symbolTable.Outer
	return instructions
}

func (c *Compiler) replaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))
	c.scopes[c.scopeIndex].lastInstruction.Opcode = code.OpReturnValue
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		c.emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		c.emit(code.OpGetBuiltin, s.Index)
	case FreeScope:
		c.emit(code.OpGetFree, s.Index)
	case FunctionScope:
		c.emit(code.OpCurrentClosure)
	}
}
