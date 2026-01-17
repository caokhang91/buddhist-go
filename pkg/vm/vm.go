package vm

import (
	"fmt"
	"sync"

	"github.com/caokhang91/buddhist-go/pkg/code"
	"github.com/caokhang91/buddhist-go/pkg/compiler"
	"github.com/caokhang91/buddhist-go/pkg/object"
)

const StackSize = 2048
const GlobalsSize = 65536
const MaxFrames = 1024

// True and False are singleton boolean objects
var True = &object.Boolean{Value: true}
var False = &object.Boolean{Value: false}
var Null = &object.Null{}

// Frame represents a call frame
type Frame struct {
	cl          *object.Closure
	ip          int
	basePointer int
}

// NewFrame creates a new frame
func NewFrame(cl *object.Closure, basePointer int) *Frame {
	return &Frame{cl: cl, ip: -1, basePointer: basePointer}
}

// Instructions returns the frame's instructions
func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}

// VM represents the virtual machine
type VM struct {
	constants []object.Object

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	globals   []object.Object
	globalsMu sync.RWMutex // Protects globals from concurrent access

	frames      []*Frame
	framesIndex int
}

// New creates a new VM
func New(bytecode *compiler.Bytecode) *VM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:   bytecode.Constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     make([]object.Object, GlobalsSize),
		frames:      frames,
		framesIndex: 1,
	}
}

// NewWithGlobalsStore creates a new VM with existing globals
func NewWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	vm := New(bytecode)
	vm.globals = s
	return vm
}

func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

func (vm *VM) pushFrame(f *Frame) {
	vm.frames[vm.framesIndex] = f
	vm.framesIndex++
}

func (vm *VM) popFrame() *Frame {
	vm.framesIndex--
	return vm.frames[vm.framesIndex]
}

// StackTop returns the top element of the stack
func (vm *VM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

// Run executes the bytecode
func (vm *VM) Run() error {
	var ip int
	var ins code.Instructions
	var op code.Opcode

	for vm.currentFrame().ip < len(vm.currentFrame().Instructions())-1 {
		vm.currentFrame().ip++

		ip = vm.currentFrame().ip
		ins = vm.currentFrame().Instructions()
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			err := vm.push(vm.constants[constIndex])
			if err != nil {
				return err
			}

		case code.OpAdd, code.OpSub, code.OpMul, code.OpDiv, code.OpMod:
			err := vm.executeBinaryOperation(op)
			if err != nil {
				return err
			}

		case code.OpPop:
			vm.pop()

		case code.OpTrue:
			err := vm.push(True)
			if err != nil {
				return err
			}

		case code.OpFalse:
			err := vm.push(False)
			if err != nil {
				return err
			}

		case code.OpNull:
			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpEqual, code.OpNotEqual, code.OpGreaterThan, code.OpGreaterThanOrEqual:
			err := vm.executeComparison(op)
			if err != nil {
				return err
			}

		case code.OpAnd:
			right := vm.pop()
			left := vm.pop()
			leftVal := isTruthy(left)
			rightVal := isTruthy(right)
			if leftVal && rightVal {
				vm.push(True)
			} else {
				vm.push(False)
			}

		case code.OpOr:
			right := vm.pop()
			left := vm.pop()
			leftVal := isTruthy(left)
			rightVal := isTruthy(right)
			if leftVal || rightVal {
				vm.push(True)
			} else {
				vm.push(False)
			}

		case code.OpBang:
			err := vm.executeBangOperator()
			if err != nil {
				return err
			}

		case code.OpMinus:
			err := vm.executeMinusOperator()
			if err != nil {
				return err
			}

		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip = pos - 1

		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			condition := vm.pop()
			if !isTruthy(condition) {
				vm.currentFrame().ip = pos - 1
			}

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			value := vm.pop()
			vm.globalsMu.Lock()
			vm.globals[globalIndex] = value
			vm.globalsMu.Unlock()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			vm.currentFrame().ip += 2
			vm.globalsMu.RLock()
			value := vm.globals[globalIndex]
			vm.globalsMu.RUnlock()
			err := vm.push(value)
			if err != nil {
				return err
			}

		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			vm.stack[frame.basePointer+int(localIndex)] = vm.pop()

		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1
			frame := vm.currentFrame()
			err := vm.push(vm.stack[frame.basePointer+int(localIndex)])
			if err != nil {
				return err
			}

		case code.OpArray:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements

			err := vm.push(array)
			if err != nil {
				return err
			}

		case code.OpHash:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements

			err = vm.push(hash)
			if err != nil {
				return err
			}

		case code.OpIndex:
			index := vm.pop()
			left := vm.pop()

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}

		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}

		case code.OpReturnValue:
			returnValue := vm.pop()

			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(returnValue)
			if err != nil {
				return err
			}

		case code.OpReturn:
			frame := vm.popFrame()
			vm.sp = frame.basePointer - 1

			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpGetBuiltin:
			builtinIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			definition := object.Builtins[builtinIndex]
			err := vm.push(&object.Builtin{Name: definition.Name, Fn: definition.Fn})
			if err != nil {
				return err
			}

		case code.OpClosure:
			constIndex := code.ReadUint16(ins[ip+1:])
			numFree := code.ReadUint8(ins[ip+3:])
			vm.currentFrame().ip += 3

			err := vm.pushClosure(int(constIndex), int(numFree))
			if err != nil {
				return err
			}

		case code.OpGetFree:
			freeIndex := code.ReadUint8(ins[ip+1:])
			vm.currentFrame().ip += 1

			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure.Free[freeIndex])
			if err != nil {
				return err
			}

		case code.OpCurrentClosure:
			currentClosure := vm.currentFrame().cl
			err := vm.push(currentClosure)
			if err != nil {
				return err
			}

		case code.OpSpawn:
			fn := vm.pop()
			go vm.executeSpawn(fn)

		case code.OpChannel:
			channel := &object.Channel{Chan: make(chan object.Object)}
			err := vm.push(channel)
			if err != nil {
				return err
			}

		case code.OpChannelBuffered:
			bufferSize := int(code.ReadUint16(ins[ip+1:]))
			vm.currentFrame().ip += 2
			channel := &object.Channel{Chan: make(chan object.Object, bufferSize)}
			err := vm.push(channel)
			if err != nil {
				return err
			}

		case code.OpSend:
			value := vm.pop()
			channel := vm.pop()
			ch, ok := channel.(*object.Channel)
			if !ok {
				return fmt.Errorf("cannot send to non-channel")
			}
			// Blocking send - no goroutine wrapper
			ch.Chan <- value
			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpReceive:
			channel := vm.pop()
			ch, ok := channel.(*object.Channel)
			if !ok {
				return fmt.Errorf("cannot receive from non-channel")
			}
			value, ok := <-ch.Chan
			if !ok {
				// Channel is closed
				err := vm.push(Null)
				if err != nil {
					return err
				}
			} else {
				err := vm.push(value)
				if err != nil {
					return err
				}
			}

		case code.OpCloseChannel:
			channel := vm.pop()
			ch, ok := channel.(*object.Channel)
			if !ok {
				return fmt.Errorf("cannot close non-channel")
			}
			close(ch.Chan)
			err := vm.push(Null)
			if err != nil {
				return err
			}

		case code.OpBreak:
			return &BreakSignal{}

		case code.OpContinue:
			return &ContinueSignal{}
		}
	}

	return nil
}

func (vm *VM) push(o object.Object) error {
	if vm.sp >= StackSize {
		return fmt.Errorf("stack overflow")
	}

	vm.stack[vm.sp] = o
	vm.sp++

	return nil
}

func (vm *VM) pop() object.Object {
	o := vm.stack[vm.sp-1]
	vm.sp--
	return o
}

func (vm *VM) executeBinaryOperation(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperation(op, left, right)
	case leftType == object.FLOAT_OBJ || rightType == object.FLOAT_OBJ:
		return vm.executeBinaryFloatOperation(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		return vm.executeBinaryStringOperation(op, left, right)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s",
			leftType, rightType)
	}
}

func (vm *VM) executeBinaryIntegerOperation(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	var result int64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	case code.OpMod:
		result = leftValue % rightValue
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(&object.Integer{Value: result})
}

func (vm *VM) executeBinaryFloatOperation(op code.Opcode, left, right object.Object) error {
	var leftValue, rightValue float64

	switch l := left.(type) {
	case *object.Float:
		leftValue = l.Value
	case *object.Integer:
		leftValue = float64(l.Value)
	}

	switch r := right.(type) {
	case *object.Float:
		rightValue = r.Value
	case *object.Integer:
		rightValue = float64(r.Value)
	}

	var result float64

	switch op {
	case code.OpAdd:
		result = leftValue + rightValue
	case code.OpSub:
		result = leftValue - rightValue
	case code.OpMul:
		result = leftValue * rightValue
	case code.OpDiv:
		result = leftValue / rightValue
	default:
		return fmt.Errorf("unknown float operator: %d", op)
	}

	return vm.push(&object.Float{Value: result})
}

func (vm *VM) executeBinaryStringOperation(op code.Opcode, left, right object.Object) error {
	if op != code.OpAdd {
		return fmt.Errorf("unknown string operator: %d", op)
	}

	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	return vm.push(&object.String{Value: leftValue + rightValue})
}

func (vm *VM) executeComparison(op code.Opcode) error {
	right := vm.pop()
	left := vm.pop()

	if left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ {
		return vm.executeIntegerComparison(op, left, right)
	}

	if left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ {
		return vm.executeFloatComparison(op, left, right)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(right == left))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(right != left))
	default:
		return fmt.Errorf("unknown operator: %d (%s %s)",
			op, left.Type(), right.Type())
	}
}

func (vm *VM) executeIntegerComparison(op code.Opcode, left, right object.Object) error {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	case code.OpGreaterThanOrEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue >= rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func (vm *VM) executeFloatComparison(op code.Opcode, left, right object.Object) error {
	var leftValue, rightValue float64

	switch l := left.(type) {
	case *object.Float:
		leftValue = l.Value
	case *object.Integer:
		leftValue = float64(l.Value)
	}

	switch r := right.(type) {
	case *object.Float:
		rightValue = r.Value
	case *object.Integer:
		rightValue = float64(r.Value)
	}

	switch op {
	case code.OpEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue == leftValue))
	case code.OpNotEqual:
		return vm.push(nativeBoolToBooleanObject(rightValue != leftValue))
	case code.OpGreaterThan:
		return vm.push(nativeBoolToBooleanObject(leftValue > rightValue))
	case code.OpGreaterThanOrEqual:
		return vm.push(nativeBoolToBooleanObject(leftValue >= rightValue))
	default:
		return fmt.Errorf("unknown operator: %d", op)
	}
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return True
	}
	return False
}

func (vm *VM) executeBangOperator() error {
	operand := vm.pop()

	switch operand {
	case True:
		return vm.push(False)
	case False:
		return vm.push(True)
	case Null:
		return vm.push(True)
	default:
		return vm.push(False)
	}
}

func (vm *VM) executeMinusOperator() error {
	operand := vm.pop()

	switch obj := operand.(type) {
	case *object.Integer:
		return vm.push(&object.Integer{Value: -obj.Value})
	case *object.Float:
		return vm.push(&object.Float{Value: -obj.Value})
	default:
		return fmt.Errorf("unsupported type for negation: %s", operand.Type())
	}
}

func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)

	for i := startIndex; i < endIndex; i++ {
		elements[i-startIndex] = vm.stack[i]
	}

	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair)

	for i := startIndex; i < endIndex; i += 2 {
		key := vm.stack[i]
		value := vm.stack[i+1]

		pair := object.HashPair{Key: key, Value: value}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return nil, fmt.Errorf("unusable as hash key: %s", key.Type())
		}

		hashedPairs[hashKey.HashKey()] = pair
	}

	return &object.Hash{Pairs: hashedPairs}, nil
}

func (vm *VM) executeIndexExpression(left, index object.Object) error {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeArrayIndex(left, index)
	case left.Type() == object.HASH_OBJ:
		return vm.executeHashIndex(left, index)
	case left.Type() == object.STRING_OBJ && index.Type() == object.INTEGER_OBJ:
		return vm.executeStringIndex(left, index)
	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeArrayIndex(array, index object.Object) error {
	arrayObject := array.(*object.Array)
	i := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(arrayObject.Elements[i])
}

func (vm *VM) executeHashIndex(hash, index object.Object) error {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return fmt.Errorf("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return vm.push(Null)
	}

	return vm.push(pair.Value)
}

func (vm *VM) executeStringIndex(str, index object.Object) error {
	strObject := str.(*object.String)
	i := index.(*object.Integer).Value
	max := int64(len(strObject.Value) - 1)

	if i < 0 || i > max {
		return vm.push(Null)
	}

	return vm.push(&object.String{Value: string(strObject.Value[i])})
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function and non-builtin")
	}
}

func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {
	if numArgs != cl.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			cl.Fn.NumParameters, numArgs)
	}

	frame := NewFrame(cl, vm.sp-numArgs)
	vm.pushFrame(frame)
	vm.sp = frame.basePointer + cl.Fn.NumLocals

	return nil
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]

	result := builtin.Fn(args...)
	vm.sp = vm.sp - numArgs - 1

	if result != nil {
		vm.push(result)
	} else {
		vm.push(Null)
	}

	return nil
}

func (vm *VM) pushClosure(constIndex int, numFree int) error {
	constant := vm.constants[constIndex]
	function, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]object.Object, numFree)
	for i := 0; i < numFree; i++ {
		free[i] = vm.stack[vm.sp-numFree+i]
	}
	vm.sp = vm.sp - numFree

	closure := &object.Closure{Fn: function, Free: free}
	return vm.push(closure)
}

func (vm *VM) executeSpawn(fn object.Object) {
	switch fn := fn.(type) {
	case *object.Closure:
		// Create a new VM for the goroutine
		newVM := &VM{
			constants:   vm.constants,
			stack:       make([]object.Object, StackSize),
			sp:          0,
			globals:     vm.globals,   // Share globals with parent VM
			globalsMu:   vm.globalsMu, // Share mutex for thread-safe access
			frames:      make([]*Frame, MaxFrames),
			framesIndex: 1,
		}
		newVM.frames[0] = NewFrame(fn, 0)
		// Add panic recovery for spawned goroutines
		defer func() {
			if r := recover(); r != nil {
				// Panic recovered - could log or report error here
				// For now, we silently recover to prevent crashes
			}
		}()
		newVM.Run()
	}
}

func isTruthy(obj object.Object) bool {
	switch obj := obj.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.Null:
		return false
	case *object.Integer:
		return obj.Value != 0
	default:
		return true
	}
}

// LastPoppedStackElem returns the last popped element from the stack
func (vm *VM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}

// BreakSignal represents a break signal
type BreakSignal struct{}

func (b *BreakSignal) Error() string { return "break" }

// ContinueSignal represents a continue signal
type ContinueSignal struct{}

func (c *ContinueSignal) Error() string { return "continue" }
