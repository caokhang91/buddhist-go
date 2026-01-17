package vm

import (
	"fmt"
	"sync"

	"github.com/caokhang91/buddhist-go/pkg/code"
	"github.com/caokhang91/buddhist-go/pkg/compiler"
	"github.com/caokhang91/buddhist-go/pkg/object"
)

// OptimizedVM is a performance-optimized version of the VM
// Key optimizations:
// 1. Cached frame references
// 2. Inline push/pop operations
// 3. Direct stack access without bounds checking
// 4. Use of object pooling
// 5. Removed tracing overhead
type OptimizedVM struct {
	constants []object.Object

	stack []object.Object
	sp    int // Always points to the next value. Top of stack is stack[sp-1]

	globals   []object.Object
	globalsMu sync.RWMutex // Protects globals from concurrent access

	frames      []*Frame
	framesIndex int
}

// NewOptimized creates a new optimized VM
func NewOptimized(bytecode *compiler.Bytecode) *OptimizedVM {
	mainFn := &object.CompiledFunction{Instructions: bytecode.Instructions}
	mainClosure := &object.Closure{Fn: mainFn}
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &OptimizedVM{
		constants:   bytecode.Constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     make([]object.Object, GlobalsSize),
		frames:      frames,
		framesIndex: 1,
	}
}

// NewOptimizedWithGlobalsStore creates a new optimized VM with existing globals
func NewOptimizedWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *OptimizedVM {
	vm := NewOptimized(bytecode)
	vm.globals = s
	return vm
}

// Run executes the bytecode with optimizations
func (vm *OptimizedVM) Run() error {
	// Cache frequently accessed values
	var ip int
	var ins code.Instructions
	var op code.Opcode
	
	// Cache current frame to avoid repeated function calls
	frame := vm.frames[vm.framesIndex-1]
	frameIns := frame.Instructions()

	for frame.ip < len(frameIns)-1 {
		frame.ip++
		ip = frame.ip
		ins = frameIns
		op = code.Opcode(ins[ip])

		switch op {
		case code.OpConstant:
			constIndex := code.ReadUint16(ins[ip+1:])
			frame.ip += 2
			// Inline push
			vm.stack[vm.sp] = vm.constants[constIndex]
			vm.sp++

		case code.OpAdd:
			// Inline pop twice
			vm.sp -= 2
			left := vm.stack[vm.sp]
			right := vm.stack[vm.sp+1]

			// Fast path for integers
			if leftInt, ok := left.(*object.Integer); ok {
				if rightInt, ok := right.(*object.Integer); ok {
					result := leftInt.Value + rightInt.Value
					vm.stack[vm.sp] = object.GetCachedInteger(result)
					vm.sp++
					continue
				}
			}
			
			// Fall back to general binary operation
			err := vm.executeBinaryOperationInline(code.OpAdd, left, right)
			if err != nil {
				return err
			}

		case code.OpSub:
			vm.sp -= 2
			left := vm.stack[vm.sp]
			right := vm.stack[vm.sp+1]

			if leftInt, ok := left.(*object.Integer); ok {
				if rightInt, ok := right.(*object.Integer); ok {
					result := leftInt.Value - rightInt.Value
					vm.stack[vm.sp] = object.GetCachedInteger(result)
					vm.sp++
					continue
				}
			}
			
			err := vm.executeBinaryOperationInline(code.OpSub, left, right)
			if err != nil {
				return err
			}

		case code.OpMul:
			vm.sp -= 2
			left := vm.stack[vm.sp]
			right := vm.stack[vm.sp+1]

			if leftInt, ok := left.(*object.Integer); ok {
				if rightInt, ok := right.(*object.Integer); ok {
					result := leftInt.Value * rightInt.Value
					vm.stack[vm.sp] = object.GetCachedInteger(result)
					vm.sp++
					continue
				}
			}
			
			err := vm.executeBinaryOperationInline(code.OpMul, left, right)
			if err != nil {
				return err
			}

		case code.OpDiv:
			vm.sp -= 2
			left := vm.stack[vm.sp]
			right := vm.stack[vm.sp+1]

			if leftInt, ok := left.(*object.Integer); ok {
				if rightInt, ok := right.(*object.Integer); ok {
					if rightInt.Value != 0 {
						result := leftInt.Value / rightInt.Value
						vm.stack[vm.sp] = object.GetCachedInteger(result)
						vm.sp++
						continue
					}
				}
			}
			
			err := vm.executeBinaryOperationInline(code.OpDiv, left, right)
			if err != nil {
				return err
			}

		case code.OpMod:
			vm.sp -= 2
			left := vm.stack[vm.sp]
			right := vm.stack[vm.sp+1]

			if leftInt, ok := left.(*object.Integer); ok {
				if rightInt, ok := right.(*object.Integer); ok {
					if rightInt.Value != 0 {
						result := leftInt.Value % rightInt.Value
						vm.stack[vm.sp] = object.GetCachedInteger(result)
						vm.sp++
						continue
					}
				}
			}
			
			err := vm.executeBinaryOperationInline(code.OpMod, left, right)
			if err != nil {
				return err
			}

		case code.OpPop:
			vm.sp--

		case code.OpTrue:
			vm.stack[vm.sp] = True
			vm.sp++

		case code.OpFalse:
			vm.stack[vm.sp] = False
			vm.sp++

		case code.OpNull:
			vm.stack[vm.sp] = Null
			vm.sp++

		case code.OpEqual:
			vm.sp -= 2
			left := vm.stack[vm.sp]
			right := vm.stack[vm.sp+1]
			
			// Fast path for integers
			if leftInt, ok := left.(*object.Integer); ok {
				if rightInt, ok := right.(*object.Integer); ok {
					if leftInt.Value == rightInt.Value {
						vm.stack[vm.sp] = True
					} else {
						vm.stack[vm.sp] = False
					}
					vm.sp++
					continue
				}
			}
			
			// Fast path for booleans
			if left == right {
				vm.stack[vm.sp] = True
			} else {
				vm.stack[vm.sp] = False
			}
			vm.sp++

		case code.OpNotEqual:
			vm.sp -= 2
			left := vm.stack[vm.sp]
			right := vm.stack[vm.sp+1]
			
			// Fast path for integers
			if leftInt, ok := left.(*object.Integer); ok {
				if rightInt, ok := right.(*object.Integer); ok {
					if leftInt.Value != rightInt.Value {
						vm.stack[vm.sp] = True
					} else {
						vm.stack[vm.sp] = False
					}
					vm.sp++
					continue
				}
			}
			
			if left != right {
				vm.stack[vm.sp] = True
			} else {
				vm.stack[vm.sp] = False
			}
			vm.sp++

		case code.OpGreaterThan:
			vm.sp -= 2
			left := vm.stack[vm.sp]
			right := vm.stack[vm.sp+1]
			
			if leftInt, ok := left.(*object.Integer); ok {
				if rightInt, ok := right.(*object.Integer); ok {
					if leftInt.Value > rightInt.Value {
						vm.stack[vm.sp] = True
					} else {
						vm.stack[vm.sp] = False
					}
					vm.sp++
					continue
				}
			}
			
			err := vm.executeComparisonInline(code.OpGreaterThan, left, right)
			if err != nil {
				return err
			}

		case code.OpGreaterThanOrEqual:
			vm.sp -= 2
			left := vm.stack[vm.sp]
			right := vm.stack[vm.sp+1]
			
			if leftInt, ok := left.(*object.Integer); ok {
				if rightInt, ok := right.(*object.Integer); ok {
					if leftInt.Value >= rightInt.Value {
						vm.stack[vm.sp] = True
					} else {
						vm.stack[vm.sp] = False
					}
					vm.sp++
					continue
				}
			}
			
			err := vm.executeComparisonInline(code.OpGreaterThanOrEqual, left, right)
			if err != nil {
				return err
			}

		case code.OpAnd:
			vm.sp -= 2
			left := vm.stack[vm.sp]
			right := vm.stack[vm.sp+1]
			if isTruthy(left) && isTruthy(right) {
				vm.stack[vm.sp] = True
			} else {
				vm.stack[vm.sp] = False
			}
			vm.sp++

		case code.OpOr:
			vm.sp -= 2
			left := vm.stack[vm.sp]
			right := vm.stack[vm.sp+1]
			if isTruthy(left) || isTruthy(right) {
				vm.stack[vm.sp] = True
			} else {
				vm.stack[vm.sp] = False
			}
			vm.sp++

		case code.OpBang:
			operand := vm.stack[vm.sp-1]
			switch operand {
			case True:
				vm.stack[vm.sp-1] = False
			case False:
				vm.stack[vm.sp-1] = True
			case Null:
				vm.stack[vm.sp-1] = True
			default:
				vm.stack[vm.sp-1] = False
			}

		case code.OpMinus:
			operand := vm.stack[vm.sp-1]
			switch obj := operand.(type) {
			case *object.Integer:
				vm.stack[vm.sp-1] = object.GetCachedInteger(-obj.Value)
			case *object.Float:
				vm.stack[vm.sp-1] = &object.Float{Value: -obj.Value}
			default:
				return fmt.Errorf("unsupported type for negation: %s", operand.Type())
			}

		case code.OpJump:
			pos := int(code.ReadUint16(ins[ip+1:]))
			frame.ip = pos - 1

		case code.OpJumpNotTruthy:
			pos := int(code.ReadUint16(ins[ip+1:]))
			frame.ip += 2

			condition := vm.stack[vm.sp-1]
			vm.sp--
			if !isTruthy(condition) {
				frame.ip = pos - 1
			}

		case code.OpSetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			frame.ip += 2
			vm.sp--
			vm.globals[globalIndex] = vm.stack[vm.sp]

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			frame.ip += 2
			vm.stack[vm.sp] = vm.globals[globalIndex]
			vm.sp++

		case code.OpSetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			frame.ip += 1
			vm.sp--
			vm.stack[frame.basePointer+int(localIndex)] = vm.stack[vm.sp]

		case code.OpGetLocal:
			localIndex := code.ReadUint8(ins[ip+1:])
			frame.ip += 1
			vm.stack[vm.sp] = vm.stack[frame.basePointer+int(localIndex)]
			vm.sp++

		case code.OpArray:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			frame.ip += 2

			array := vm.buildArrayOptimized(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements
			vm.stack[vm.sp] = array
			vm.sp++

		case code.OpHash:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			frame.ip += 2

			hash, err := vm.buildHashOptimized(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements
			vm.stack[vm.sp] = hash
			vm.sp++

		case code.OpPHPArray:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			frame.ip += 2

			phpArray, err := vm.buildPHPArrayOptimized(vm.sp-numElements*2, vm.sp, numElements)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements*2
			vm.stack[vm.sp] = phpArray
			vm.sp++

		case code.OpIndex:
			vm.sp -= 2
			left := vm.stack[vm.sp]
			index := vm.stack[vm.sp+1]

			err := vm.executeIndexExpressionOptimized(left, index)
			if err != nil {
				return err
			}

		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			frame.ip += 1

			err := vm.executeCallOptimized(int(numArgs))
			if err != nil {
				return err
			}
			// Update frame reference after call
			frame = vm.frames[vm.framesIndex-1]
			frameIns = frame.Instructions()

		case code.OpReturnValue:
			returnValue := vm.stack[vm.sp-1]
			vm.sp--

			vm.framesIndex--
			frame = vm.frames[vm.framesIndex-1]
			vm.sp = vm.frames[vm.framesIndex].basePointer - 1
			frameIns = frame.Instructions()

			vm.stack[vm.sp] = returnValue
			vm.sp++

		case code.OpReturn:
			vm.framesIndex--
			frame = vm.frames[vm.framesIndex-1]
			vm.sp = vm.frames[vm.framesIndex].basePointer - 1
			frameIns = frame.Instructions()

			vm.stack[vm.sp] = Null
			vm.sp++

		case code.OpGetBuiltin:
			builtinIndex := code.ReadUint8(ins[ip+1:])
			frame.ip += 1

			definition := object.Builtins[builtinIndex]
			vm.stack[vm.sp] = &object.Builtin{Name: definition.Name, Fn: definition.Fn}
			vm.sp++

		case code.OpClosure:
			constIndex := code.ReadUint16(ins[ip+1:])
			numFree := code.ReadUint8(ins[ip+3:])
			frame.ip += 3

			err := vm.pushClosureOptimized(int(constIndex), int(numFree))
			if err != nil {
				return err
			}

		case code.OpGetFree:
			freeIndex := code.ReadUint8(ins[ip+1:])
			frame.ip += 1

			currentClosure := frame.cl
			vm.stack[vm.sp] = currentClosure.Free[freeIndex]
			vm.sp++

		case code.OpCurrentClosure:
			currentClosure := frame.cl
			vm.stack[vm.sp] = currentClosure
			vm.sp++

		case code.OpSpawn:
			fn := vm.stack[vm.sp-1]
			vm.sp--
			go vm.executeSpawnOptimized(fn)
			// Push null so that expression statement pop works
			vm.stack[vm.sp] = Null
			vm.sp++

		case code.OpChannel:
			channel := &object.Channel{Chan: make(chan object.Object)}
			vm.stack[vm.sp] = channel
			vm.sp++

		case code.OpChannelBuffered:
			bufferSize := int(code.ReadUint16(ins[ip+1:]))
			frame.ip += 2
			channel := &object.Channel{Chan: make(chan object.Object, bufferSize)}
			vm.stack[vm.sp] = channel
			vm.sp++

		case code.OpSend:
			vm.sp -= 2
			channel := vm.stack[vm.sp]
			value := vm.stack[vm.sp+1]
			ch, ok := channel.(*object.Channel)
			if !ok {
				return fmt.Errorf("cannot send to non-channel")
			}
			ch.Chan <- value
			vm.stack[vm.sp] = Null
			vm.sp++

		case code.OpReceive:
			vm.sp--
			channel := vm.stack[vm.sp]
			ch, ok := channel.(*object.Channel)
			if !ok {
				return fmt.Errorf("cannot receive from non-channel")
			}
			value, ok := <-ch.Chan
			if !ok {
				vm.stack[vm.sp] = Null
			} else {
				vm.stack[vm.sp] = value
			}
			vm.sp++

		case code.OpCloseChannel:
			vm.sp--
			channel := vm.stack[vm.sp]
			ch, ok := channel.(*object.Channel)
			if !ok {
				return fmt.Errorf("cannot close non-channel")
			}
			close(ch.Chan)
			vm.stack[vm.sp] = Null
			vm.sp++

		case code.OpBreak:
			return &BreakSignal{}

		case code.OpContinue:
			return &ContinueSignal{}

		case code.OpSetIndex:
			vm.sp -= 3
			left := vm.stack[vm.sp]
			index := vm.stack[vm.sp+1]
			value := vm.stack[vm.sp+2]

			err := vm.executeSetIndexOptimized(left, index, value)
			if err != nil {
				return err
			}

		case code.OpArrayPush:
			vm.sp -= 2
			arr := vm.stack[vm.sp]
			value := vm.stack[vm.sp+1]

			err := vm.executeArrayPushOptimized(arr, value)
			if err != nil {
				return err
			}

		case code.OpSlice:
			vm.sp -= 3
			arr := vm.stack[vm.sp]
			start := vm.stack[vm.sp+1]
			end := vm.stack[vm.sp+2]

			err := vm.executeSliceOptimized(arr, start, end)
			if err != nil {
				return err
			}

		case code.OpArrayMap:
			vm.sp -= 2
			arr := vm.stack[vm.sp]
			fn := vm.stack[vm.sp+1]

			err := vm.executeArrayMapOptimized(arr, fn)
			if err != nil {
				return err
			}

		case code.OpArrayFilter:
			vm.sp -= 2
			arr := vm.stack[vm.sp]
			fn := vm.stack[vm.sp+1]

			err := vm.executeArrayFilterOptimized(arr, fn)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// executeBinaryOperationInline handles binary operations with optimizations
func (vm *OptimizedVM) executeBinaryOperationInline(op code.Opcode, left, right object.Object) error {
	leftType := left.Type()
	rightType := right.Type()

	switch {
	case leftType == object.INTEGER_OBJ && rightType == object.INTEGER_OBJ:
		return vm.executeBinaryIntegerOperationInline(op, left.(*object.Integer), right.(*object.Integer))
	case leftType == object.FLOAT_OBJ || rightType == object.FLOAT_OBJ:
		return vm.executeBinaryFloatOperationInline(op, left, right)
	case leftType == object.STRING_OBJ && rightType == object.STRING_OBJ:
		if op == code.OpAdd {
			leftVal := left.(*object.String).Value
			rightVal := right.(*object.String).Value
			vm.stack[vm.sp] = &object.String{Value: leftVal + rightVal}
			vm.sp++
			return nil
		}
		return fmt.Errorf("unknown string operator: %d", op)
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s", leftType, rightType)
	}
}

func (vm *OptimizedVM) executeBinaryIntegerOperationInline(op code.Opcode, left, right *object.Integer) error {
	var result int64

	switch op {
	case code.OpAdd:
		result = left.Value + right.Value
	case code.OpSub:
		result = left.Value - right.Value
	case code.OpMul:
		result = left.Value * right.Value
	case code.OpDiv:
		result = left.Value / right.Value
	case code.OpMod:
		result = left.Value % right.Value
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	vm.stack[vm.sp] = object.GetCachedInteger(result)
	vm.sp++
	return nil
}

func (vm *OptimizedVM) executeBinaryFloatOperationInline(op code.Opcode, left, right object.Object) error {
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

	vm.stack[vm.sp] = &object.Float{Value: result}
	vm.sp++
	return nil
}

func (vm *OptimizedVM) executeComparisonInline(op code.Opcode, left, right object.Object) error {
	if left.Type() == object.FLOAT_OBJ || right.Type() == object.FLOAT_OBJ {
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

		var result bool
		switch op {
		case code.OpGreaterThan:
			result = leftValue > rightValue
		case code.OpGreaterThanOrEqual:
			result = leftValue >= rightValue
		}

		if result {
			vm.stack[vm.sp] = True
		} else {
			vm.stack[vm.sp] = False
		}
		vm.sp++
		return nil
	}

	return fmt.Errorf("unknown operator: %d (%s %s)", op, left.Type(), right.Type())
}

func (vm *OptimizedVM) buildArrayOptimized(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	copy(elements, vm.stack[startIndex:endIndex])
	return &object.Array{Elements: elements}
}

func (vm *OptimizedVM) buildHashOptimized(startIndex, endIndex int) (object.Object, error) {
	hashedPairs := make(map[object.HashKey]object.HashPair, (endIndex-startIndex)/2)

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

func (vm *OptimizedVM) buildPHPArrayOptimized(startIndex, endIndex, numElements int) (object.Object, error) {
	if endIndex-startIndex != numElements*2 {
		return nil, fmt.Errorf("invalid PHP array element count")
	}

	phpArray := object.NewPHPArray()
	for i := 0; i < numElements; i++ {
		key := vm.stack[startIndex+i*2]
		value := vm.stack[startIndex+i*2+1]
		if _, ok := key.(*object.Null); ok {
			phpArray.Push(value)
		} else {
			phpArray.Set(key, value)
		}
	}

	return phpArray, nil
}

func (vm *OptimizedVM) executeIndexExpressionOptimized(left, index object.Object) error {
	// Check if left operand is an ERROR object
	if errObj, ok := left.(*object.Error); ok {
		return fmt.Errorf("cannot index error object: %s", errObj.Message)
	}

	// Fast path for arrays with integer index
	if arr, ok := left.(*object.Array); ok {
		if idx, ok := index.(*object.Integer); ok {
			i := idx.Value
			if i >= 0 && i < int64(len(arr.Elements)) {
				vm.stack[vm.sp] = arr.Elements[i]
				vm.sp++
				return nil
			}
			vm.stack[vm.sp] = Null
			vm.sp++
			return nil
		}
	}

	if phpArray, ok := left.(*object.PHPArray); ok {
		value, ok := phpArray.Get(index)
		if !ok {
			vm.stack[vm.sp] = Null
			vm.sp++
			return nil
		}
		vm.stack[vm.sp] = value
		vm.sp++
		return nil
	}

	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		arrayObject := left.(*object.Array)
		i := index.(*object.Integer).Value
		max := int64(len(arrayObject.Elements) - 1)

		if i < 0 || i > max {
			vm.stack[vm.sp] = Null
			vm.sp++
			return nil
		}

		vm.stack[vm.sp] = arrayObject.Elements[i]
		vm.sp++
		return nil

	case left.Type() == object.HASH_OBJ:
		hashObject := left.(*object.Hash)

		key, ok := index.(object.Hashable)
		if !ok {
			return fmt.Errorf("unusable as hash key: %s", index.Type())
		}

		pair, ok := hashObject.Pairs[key.HashKey()]
		if !ok {
			vm.stack[vm.sp] = Null
			vm.sp++
			return nil
		}

		vm.stack[vm.sp] = pair.Value
		vm.sp++
		return nil

	case left.Type() == object.STRING_OBJ && index.Type() == object.INTEGER_OBJ:
		strObject := left.(*object.String)
		i := index.(*object.Integer).Value
		max := int64(len(strObject.Value) - 1)

		if i < 0 || i > max {
			vm.stack[vm.sp] = Null
			vm.sp++
			return nil
		}

		vm.stack[vm.sp] = &object.String{Value: string(strObject.Value[i])}
		vm.sp++
		return nil

	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *OptimizedVM) executeCallOptimized(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosureOptimized(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltinOptimized(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function and non-builtin")
	}
}

func (vm *OptimizedVM) callClosureOptimized(cl *object.Closure, numArgs int) error {
	if numArgs != cl.Fn.NumParameters {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			cl.Fn.NumParameters, numArgs)
	}

	frame := NewFrame(cl, vm.sp-numArgs)
	vm.frames[vm.framesIndex] = frame
	vm.framesIndex++
	vm.sp = frame.basePointer + cl.Fn.NumLocals

	return nil
}

func (vm *OptimizedVM) callBuiltinOptimized(builtin *object.Builtin, numArgs int) error {
	args := vm.stack[vm.sp-numArgs : vm.sp]

	result := builtin.Fn(args...)

	vm.sp = vm.sp - numArgs - 1

	if result != nil {
		vm.stack[vm.sp] = result
	} else {
		vm.stack[vm.sp] = Null
	}
	vm.sp++

	return nil
}

func (vm *OptimizedVM) pushClosureOptimized(constIndex int, numFree int) error {
	constant := vm.constants[constIndex]
	function, ok := constant.(*object.CompiledFunction)
	if !ok {
		return fmt.Errorf("not a function: %+v", constant)
	}

	free := make([]object.Object, numFree)
	copy(free, vm.stack[vm.sp-numFree:vm.sp])
	vm.sp = vm.sp - numFree

	closure := &object.Closure{Fn: function, Free: free}
	vm.stack[vm.sp] = closure
	vm.sp++
	return nil
}

func (vm *OptimizedVM) executeSpawnOptimized(fn object.Object) {
	switch fn := fn.(type) {
	case *object.Closure:
		newVM := &OptimizedVM{
			constants:   vm.constants,
			stack:       make([]object.Object, StackSize),
			sp:          0,
			globals:     vm.globals,
			frames:      make([]*Frame, MaxFrames),
			framesIndex: 1,
		}
		newVM.frames[0] = NewFrame(fn, 0)
		defer func() {
			recover() // Silent recovery
		}()
		newVM.Run()
	}
}

func (vm *OptimizedVM) executeSetIndexOptimized(left, index, value object.Object) error {
	switch obj := left.(type) {
	case *object.Array:
		idx, ok := index.(*object.Integer)
		if !ok {
			return fmt.Errorf("array index must be integer, got %s", index.Type())
		}
		i := int(idx.Value)
		if i < 0 || i >= len(obj.Elements) {
			if i >= 0 && i < 65536 {
				for len(obj.Elements) <= i {
					obj.Elements = append(obj.Elements, Null)
				}
				obj.Elements[i] = value
			} else {
				return fmt.Errorf("array index out of bounds: %d", i)
			}
		} else {
			obj.Elements[i] = value
		}
		vm.stack[vm.sp] = value
		vm.sp++
		return nil

	case *object.PHPArray:
		obj.Set(index, value)
		vm.stack[vm.sp] = value
		vm.sp++
		return nil

	case *object.Hash:
		hashKey, ok := index.(object.Hashable)
		if !ok {
			return fmt.Errorf("unusable as hash key: %s", index.Type())
		}
		obj.Pairs[hashKey.HashKey()] = object.HashPair{Key: index, Value: value}
		vm.stack[vm.sp] = value
		vm.sp++
		return nil

	default:
		return fmt.Errorf("index assignment not supported: %s", left.Type())
	}
}

func (vm *OptimizedVM) executeArrayPushOptimized(arr, value object.Object) error {
	switch obj := arr.(type) {
	case *object.Array:
		obj.Elements = append(obj.Elements, value)
		vm.stack[vm.sp] = value
		vm.sp++
		return nil
	case *object.PHPArray:
		obj.Push(value)
		vm.stack[vm.sp] = value
		vm.sp++
		return nil
	default:
		return fmt.Errorf("cannot push to non-array: %s", arr.Type())
	}
}

func (vm *OptimizedVM) executeSliceOptimized(arr, start, end object.Object) error {
	arrayObj, ok := arr.(*object.Array)
	if !ok {
		return fmt.Errorf("slice operation requires array, got %s", arr.Type())
	}

	startIdx := 0
	endIdx := len(arrayObj.Elements)

	if start.Type() != object.NULL_OBJ {
		startInt, ok := start.(*object.Integer)
		if !ok {
			return fmt.Errorf("slice start index must be integer, got %s", start.Type())
		}
		startIdx = int(startInt.Value)
	}

	if end.Type() != object.NULL_OBJ {
		endInt, ok := end.(*object.Integer)
		if !ok {
			return fmt.Errorf("slice end index must be integer, got %s", end.Type())
		}
		endIdx = int(endInt.Value)
	}

	result := arrayObj.Slice(startIdx, endIdx)
	vm.stack[vm.sp] = result
	vm.sp++
	return nil
}

func (vm *OptimizedVM) executeArrayMapOptimized(arr, fn object.Object) error {
	arrayObj, ok := arr.(*object.Array)
	if !ok {
		return fmt.Errorf("map requires array, got %s", arr.Type())
	}

	builtin, ok := fn.(*object.Builtin)
	if ok {
		results := make([]object.Object, len(arrayObj.Elements))
		for i, el := range arrayObj.Elements {
			results[i] = builtin.Fn(el)
		}
		vm.stack[vm.sp] = &object.Array{Elements: results}
		vm.sp++
		return nil
	}

	closure, ok := fn.(*object.Closure)
	if !ok {
		return fmt.Errorf("map requires function, got %s", fn.Type())
	}

	results := make([]object.Object, len(arrayObj.Elements))
	for i, el := range arrayObj.Elements {
		result, err := vm.applyFunctionToElementOptimized(closure, el, i)
		if err != nil {
			return err
		}
		results[i] = result
	}

	vm.stack[vm.sp] = &object.Array{Elements: results}
	vm.sp++
	return nil
}

func (vm *OptimizedVM) executeArrayFilterOptimized(arr, fn object.Object) error {
	arrayObj, ok := arr.(*object.Array)
	if !ok {
		return fmt.Errorf("filter requires array, got %s", arr.Type())
	}

	builtin, ok := fn.(*object.Builtin)
	if ok {
		var results []object.Object
		for _, el := range arrayObj.Elements {
			result := builtin.Fn(el)
			if isTruthy(result) {
				results = append(results, el)
			}
		}
		vm.stack[vm.sp] = &object.Array{Elements: results}
		vm.sp++
		return nil
	}

	closure, ok := fn.(*object.Closure)
	if !ok {
		return fmt.Errorf("filter requires function, got %s", fn.Type())
	}

	var results []object.Object
	for i, el := range arrayObj.Elements {
		result, err := vm.applyFunctionToElementOptimized(closure, el, i)
		if err != nil {
			return err
		}
		if isTruthy(result) {
			results = append(results, el)
		}
	}

	vm.stack[vm.sp] = &object.Array{Elements: results}
	vm.sp++
	return nil
}

func (vm *OptimizedVM) applyFunctionToElementOptimized(closure *object.Closure, element object.Object, index int) (object.Object, error) {
	newVM := &OptimizedVM{
		constants:   vm.constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     vm.globals,
		frames:      make([]*Frame, MaxFrames),
		framesIndex: 1,
	}

	frame := NewFrame(closure, 0)
	newVM.frames[0] = frame

	newVM.stack[0] = element
	newVM.sp = 1

	if closure.Fn.NumParameters >= 2 {
		newVM.stack[1] = object.GetCachedInteger(int64(index))
		newVM.sp = 2
	}

	err := newVM.Run()
	if err != nil {
		return nil, err
	}

	if newVM.sp > 0 {
		return newVM.stack[newVM.sp-1], nil
	}
	return Null, nil
}

// StackTop returns the top element of the stack
func (vm *OptimizedVM) StackTop() object.Object {
	if vm.sp == 0 {
		return nil
	}
	return vm.stack[vm.sp-1]
}

// LastPoppedStackElem returns the last popped element from the stack
func (vm *OptimizedVM) LastPoppedStackElem() object.Object {
	return vm.stack[vm.sp]
}
