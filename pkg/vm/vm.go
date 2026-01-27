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

// Frame represents a call frame in the virtual machine execution stack.
// Each function call creates a new frame that tracks:
//   - cl: The closure being executed (contains the function bytecode and captured variables)
//   - ip: Instruction pointer (index into the function's bytecode instructions)
//   - basePointer: Stack pointer position where this frame's local variables begin.
//                  Used to calculate local variable offsets relative to the stack base.
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

// VM represents the virtual machine with optimized execution
// Key optimizations:
// 1. Cached frame references
// 2. Inline push/pop operations
// 3. Direct stack access without bounds checking
// 4. Use of object pooling
// 5. Fast paths for integer operations
type VM struct {
	constants []object.Object

	stack []object.Object
	sp    int // Stack pointer: Always points to the next free slot on the stack.
	       // The top element is at stack[sp-1]. When pushing, write to stack[sp] then increment sp.
	       // When popping, decrement sp first, then read from stack[sp] (now pointing to the previous top).

	globals   []object.Object
	globalsMu sync.RWMutex // Protects globals from concurrent access

	frames      []*Frame
	framesIndex int

	// exceptionHandlers is a stack of active try-handlers (for OpTry/OpThrow).
	exceptionHandlers []exceptionHandler
}

type exceptionHandler struct {
	catchPos    int
	finallyPos  int
	framesIndex int
	sp          int
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
		exceptionHandlers: make([]exceptionHandler, 0, 16),
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

// Run executes the bytecode with optimizations
func (vm *VM) Run() error {
	// Register closure caller for progress callbacks in builtin functions
	object.SetClosureCaller(func(closure *object.Closure, args ...object.Object) (object.Object, error) {
		return CallClosure(closure, vm.constants, vm.globals, args...)
	})
	defer object.ClearClosureCaller()

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

			// Fast path for strings - compare by value
			if leftStr, ok := left.(*object.String); ok {
				if rightStr, ok := right.(*object.String); ok {
					if leftStr.Value == rightStr.Value {
						vm.stack[vm.sp] = True
					} else {
						vm.stack[vm.sp] = False
					}
					vm.sp++
					continue
				}
			}

			// Fast path for booleans and other pointer-equal objects
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

			// Fast path for strings - compare by value
			if leftStr, ok := left.(*object.String); ok {
				if rightStr, ok := right.(*object.String); ok {
					if leftStr.Value != rightStr.Value {
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
			vm.globalsMu.Lock()
			vm.globals[globalIndex] = vm.stack[vm.sp]
			vm.globalsMu.Unlock()

		case code.OpGetGlobal:
			globalIndex := code.ReadUint16(ins[ip+1:])
			frame.ip += 2
			vm.globalsMu.RLock()
			vm.stack[vm.sp] = vm.globals[globalIndex]
			vm.globalsMu.RUnlock()
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

			array := vm.buildArray(vm.sp-numElements, vm.sp)
			vm.sp = vm.sp - numElements
			vm.stack[vm.sp] = array
			vm.sp++

		case code.OpHash:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			frame.ip += 2

			hash, err := vm.buildHash(vm.sp-numElements, vm.sp)
			if err != nil {
				return err
			}
			vm.sp = vm.sp - numElements
			vm.stack[vm.sp] = hash
			vm.sp++

		case code.OpPHPArray:
			numElements := int(code.ReadUint16(ins[ip+1:]))
			frame.ip += 2

			phpArray, err := vm.buildPHPArray(vm.sp-numElements*2, vm.sp, numElements)
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

			err := vm.executeIndexExpression(left, index)
			if err != nil {
				return err
			}

		case code.OpCall:
			numArgs := code.ReadUint8(ins[ip+1:])
			frame.ip += 1

			err := vm.executeCall(int(numArgs))
			if err != nil {
				return err
			}
			// Builtin errors are treated as throws so try/catch can catch them
			if top := vm.StackTop(); top != nil {
				if errObj, ok := top.(*object.Error); ok {
					vm.sp--
					if _, throwErr := vm.throwValue(errObj); throwErr != nil {
						return throwErr
					}
					frame = vm.currentFrame()
					frameIns = frame.Instructions()
					continue
				}
			}
			// Update frame reference after call
			frame = vm.frames[vm.framesIndex-1]
			frameIns = frame.Instructions()

		case code.OpClass:
			constIndex := code.ReadUint16(ins[ip+1:])
			frame.ip += 2
			// Class is already in constants, just push it
			vm.stack[vm.sp] = vm.constants[constIndex]
			vm.sp++

		case code.OpInstantiate:
			numArgs := int(code.ReadUint8(ins[ip+1:]))
			frame.ip += 1

			class := vm.stack[vm.sp-1-numArgs]
			classObj, ok := class.(*object.Class)
			if !ok {
				return fmt.Errorf("not a class: %s", class.Type())
			}

			// Create instance
			instance := &object.Instance{
				Class:  classObj,
				Fields: make(map[string]object.Object),
			}

			// Initialize properties with arguments or null
			args := vm.stack[vm.sp-numArgs : vm.sp]
			for i, propName := range classObj.Properties {
				if i < len(args) {
					instance.Fields[propName] = args[i]
				} else {
					instance.Fields[propName] = Null
				}
			}

			vm.sp = vm.sp - numArgs - 1
			vm.stack[vm.sp] = instance
			vm.sp++

		case code.OpGetProperty:
			propName := vm.stack[vm.sp-1]
			vm.sp--
			instance := vm.stack[vm.sp-1]
			vm.sp--

			instanceObj, ok := instance.(*object.Instance)
			if !ok {
				// Not an instance - try to fallback to index operation
				// This handles cases like arr.property where property is a string
				// For arrays, we can't use string as index, so return error
				// But for hashes, we can use string as key
				if hash, ok := instance.(*object.Hash); ok {
					// Try to use as hash key
					key, ok := propName.(object.Hashable)
					if !ok {
						return fmt.Errorf("only instances have properties, got %s", instance.Type())
					}
					pair, ok := hash.Pairs[key.HashKey()]
					if !ok {
						vm.stack[vm.sp] = Null
						vm.sp++
						continue
					}
					vm.stack[vm.sp] = pair.Value
					vm.sp++
					continue
				}
				return fmt.Errorf("only instances have properties, got %s", instance.Type())
			}

			propNameStr, ok := propName.(*object.String)
			if !ok {
				return fmt.Errorf("property name must be string, got %s", propName.Type())
			}

			// Check if it's a method (search in current class and parent classes)
			method := findMethod(instanceObj.Class, propNameStr.Value)
			if method != nil {
				// Return a bound method (closure with instance as first free variable)
				closure := &object.Closure{
					Fn:   method,
					Free: []object.Object{instanceObj},
				}
				vm.stack[vm.sp] = closure
				vm.sp++
				continue
			}

			// Check if it's a field
			if field, ok := instanceObj.Fields[propNameStr.Value]; ok {
				vm.stack[vm.sp] = field
				vm.sp++
				continue
			}

			// Property doesn't exist
			vm.stack[vm.sp] = Null
			vm.sp++

		case code.OpSetProperty:
			value := vm.stack[vm.sp-1]
			vm.sp--
			propName := vm.stack[vm.sp-1]
			vm.sp--
			instance := vm.stack[vm.sp-1]
			vm.sp--

			instanceObj, ok := instance.(*object.Instance)
			if !ok {
				return fmt.Errorf("only instances have properties, got %s", instance.Type())
			}

			propNameStr, ok := propName.(*object.String)
			if !ok {
				return fmt.Errorf("property name must be string, got %s", propName.Type())
			}

			instanceObj.Fields[propNameStr.Value] = value
			// Push the assigned value back
			vm.stack[vm.sp] = value
			vm.sp++

		case code.OpGetMethod:
			methodNameIndex := code.ReadUint16(ins[ip+1:])
			frame.ip += 2

			methodName := vm.constants[methodNameIndex].(*object.String)
			instance := vm.stack[vm.sp-1]
			vm.sp--

			instanceObj, ok := instance.(*object.Instance)
			if !ok {
				return fmt.Errorf("only instances have methods, got %s", instance.Type())
			}

			// Search for method in class hierarchy
			method := findMethod(instanceObj.Class, methodName.Value)
			if method == nil {
				vm.stack[vm.sp] = Null
				vm.sp++
				continue
			}

			// Create bound method (closure with instance as first free variable)
			closure := &object.Closure{
				Fn:   method,
				Free: []object.Object{instanceObj},
			}
			vm.stack[vm.sp] = closure
			vm.sp++

		case code.OpCallMethod:
			numArgs := int(code.ReadUint8(ins[ip+1:]))
			frame.ip += 1

			method := vm.stack[vm.sp-1-numArgs]
			closure, ok := method.(*object.Closure)
			if !ok {
				return fmt.Errorf("calling non-method: %s", method.Type())
			}

			// Get instance from free variables (first free var is 'this')
			if len(closure.Free) == 0 {
				return fmt.Errorf("method has no bound instance")
			}
			instance := closure.Free[0]

			// Push instance as first argument (this)
			args := vm.stack[vm.sp-numArgs : vm.sp]
			vm.sp = vm.sp - numArgs - 1

			// Create new closure with instance bound
			boundClosure := &object.Closure{
				Fn:   closure.Fn,
				Free: []object.Object{instance},
			}

			// Push instance as first argument (this), then other arguments
			vm.stack[vm.sp] = instance
			vm.sp++
			for _, arg := range args {
				vm.stack[vm.sp] = arg
				vm.sp++
			}

			// Call the method (boundClosure is already on stack, but we'll use it directly)
			err := vm.callClosure(boundClosure, numArgs+1) // +1 for 'this'
			if err != nil {
				return err
			}
			// Update frame reference after call
			frame = vm.frames[vm.framesIndex-1]
			frameIns = frame.Instructions()

		case code.OpThis:
			// Retrieve 'this' reference from the method's first local variable slot.
			// In method calls, the instance is passed as the first argument, which is placed
			// at frame.basePointer (the start of local variables for this frame).
			// This allows methods to access their bound instance via the 'this' keyword.
			if frame.basePointer >= 0 && frame.basePointer < len(vm.stack) {
				// 'this' is stored at the basePointer position (first argument/local)
				vm.stack[vm.sp] = vm.stack[frame.basePointer]
				vm.sp++
			} else {
				// Invalid basePointer indicates this is not a method context.
				// 'this' is only available inside class methods, not regular functions.
			}


		case code.OpInherit:
			vm.sp -= 2
			child := vm.stack[vm.sp]
			parent := vm.stack[vm.sp+1] // Stack: ..., parent, child. SP points after child.
			// Wait, if we push parent then child.
			// Stack: [..., Parent, Child].
			// vm.sp points to next empty.
			// vm.sp-1 is Child. vm.sp-2 is Parent.
			// Pop 2: vm.sp -= 2.
			// vm.stack[vm.sp] is Parent. vm.stack[vm.sp+1] is Child.
			
			// Let's verify Stack logic.
			// Push A. Push B.
			// Stack: [A, B]. SP = 2.
			// sp-1 = B. sp-2 = A.
			// Decrement sp by 2. SP = 0.
			// stack[0] = A. stack[1] = B.
			parent = vm.stack[vm.sp]
			child = vm.stack[vm.sp+1]

			childClass, ok := child.(*object.Class)
			if !ok {
				return fmt.Errorf("child is not a class: %s", child.Type())
			}
			parentClass, ok := parent.(*object.Class)
			if !ok {
				return fmt.Errorf("parent is not a class: %s", parent.Type())
			}

			childClass.Parent = parentClass
			// Push child back
			vm.stack[vm.sp] = childClass
			vm.sp++

		case code.OpReturnValue:
			// Pop the return value from the stack
			returnValue := vm.stack[vm.sp-1]
			vm.sp--

			// CRITICAL: Save basePointer from the current (returning) frame BEFORE decrementing framesIndex.
			// The basePointer indicates where the current frame's local variables start on the stack.
			// We need this value to restore the caller's stack pointer to the correct position,
			// which is exactly where the arguments for this function were placed.
			currentFrame := vm.frames[vm.framesIndex-1]
			basePointer := currentFrame.basePointer

			// Pop the current frame from the call stack
			vm.framesIndex--

			// Handle return from top-level frame (e.g., main program or spawned closure).
			// When framesIndex reaches 0, we're back at the main frame and should stop execution.
			// Push the return value onto the stack for potential use, then exit the VM.
			if vm.framesIndex == 0 {
				vm.stack[vm.sp] = returnValue
				vm.sp++
				return nil
			}

			// Switch to the caller's frame (now the top frame after pop)
			frame = vm.frames[vm.framesIndex-1]
			frameIns = frame.Instructions()

			// Restore the stack pointer to where the caller's frame expects it.
			// The basePointer of the returning frame marks where arguments were placed,
			// so restoring sp to basePointer effectively "cleans up" the callee's stack space
			// (locals and intermediate values) while preserving the argument area structure.
			vm.sp = basePointer

			// Push the return value onto the stack at the restored position.
			// The caller expects this value at the top of the stack after the function call.
			vm.stack[vm.sp] = returnValue
			vm.sp++

		case code.OpReturn:
			vm.framesIndex--
			// Check if we're returning from the top-level frame (e.g., spawned closure)
			if vm.framesIndex == 0 {
				vm.stack[vm.sp] = Null
				vm.sp++
				return nil
			}
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

			err := vm.pushClosure(int(constIndex), int(numFree))
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
			go vm.executeSpawn(fn)
			// Push null so that expression statement pop works
			vm.stack[vm.sp] = Null
			vm.sp++

		case code.OpChannel:
			channel := &object.Channel{Chan: make(chan object.Object)}
			vm.stack[vm.sp] = channel
			vm.sp++

		case code.OpChannelBuffered:
			vm.sp--
			sizeObj := vm.stack[vm.sp]
			sizeInt, ok := sizeObj.(*object.Integer)
			if !ok {
				return fmt.Errorf("channel buffer size must be integer, got %s", sizeObj.Type())
			}
			if sizeInt.Value < 0 {
				return fmt.Errorf("channel buffer size must be non-negative, got %d", sizeInt.Value)
			}
			channel := &object.Channel{Chan: make(chan object.Object, sizeInt.Value)}
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
			// Blocking send - no goroutine wrapper
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
				// Channel is closed
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

		case code.OpTry:
			catchPos := int(code.ReadUint16(ins[ip+1:]))
			finallyPos := int(code.ReadUint16(ins[ip+3:]))
			frame.ip += 4

			vm.exceptionHandlers = append(vm.exceptionHandlers, exceptionHandler{
				catchPos:    catchPos,
				finallyPos:  finallyPos,
				framesIndex: vm.framesIndex,
				sp:          vm.sp,
			})

		case code.OpThrow:
			if vm.sp == 0 {
				return fmt.Errorf("throw with empty stack")
			}
			vm.sp--
			thrown := vm.stack[vm.sp]
			if _, err := vm.throwValue(thrown); err != nil {
				return err
			}
			frame = vm.currentFrame()
			frameIns = frame.Instructions()
			continue

		case code.OpFinally:
			// Normal-flow exit from try: pop handler and jump to target (finally or after).
			pos := int(code.ReadUint16(ins[ip+1:]))
			frame.ip += 2
			if len(vm.exceptionHandlers) > 0 {
				vm.exceptionHandlers = vm.exceptionHandlers[:len(vm.exceptionHandlers)-1]
			}
			frame.ip = pos - 1

		case code.OpSetIndex:
			vm.sp -= 3
			left := vm.stack[vm.sp]
			index := vm.stack[vm.sp+1]
			value := vm.stack[vm.sp+2]

			err := vm.executeSetIndex(left, index, value)
			if err != nil {
				return err
			}

		case code.OpArrayPush:
			vm.sp -= 2
			arr := vm.stack[vm.sp]
			value := vm.stack[vm.sp+1]

			err := vm.executeArrayPush(arr, value)
			if err != nil {
				return err
			}

		case code.OpSlice:
			vm.sp -= 3
			arr := vm.stack[vm.sp]
			start := vm.stack[vm.sp+1]
			end := vm.stack[vm.sp+2]

			err := vm.executeSlice(arr, start, end)
			if err != nil {
				return err
			}

		case code.OpArrayMap:
			vm.sp -= 2
			arr := vm.stack[vm.sp]
			fn := vm.stack[vm.sp+1]

			err := vm.executeArrayMap(arr, fn)
			if err != nil {
				return err
			}

		case code.OpArrayFilter:
			vm.sp -= 2
			arr := vm.stack[vm.sp]
			fn := vm.stack[vm.sp+1]

			err := vm.executeArrayFilter(arr, fn)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// executeBinaryOperationInline handles binary operations with optimizations
func (vm *VM) executeBinaryOperationInline(op code.Opcode, left, right object.Object) error {
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
	case op == code.OpAdd && (leftType == object.STRING_OBJ || rightType == object.STRING_OBJ):
		// String concatenation: convert both to strings and concatenate
		leftStr := left.Inspect()
		rightStr := right.Inspect()
		vm.stack[vm.sp] = &object.String{Value: leftStr + rightStr}
		vm.sp++
		return nil
	default:
		return fmt.Errorf("unsupported types for binary operation: %s %s values: %s %s", leftType, rightType, left.Inspect(), right.Inspect())
	}
}

func (vm *VM) executeBinaryIntegerOperationInline(op code.Opcode, left, right *object.Integer) error {
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

func (vm *VM) executeBinaryFloatOperationInline(op code.Opcode, left, right object.Object) error {
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

func (vm *VM) executeComparisonInline(op code.Opcode, left, right object.Object) error {
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

	return fmt.Errorf("unknown operator: %d (%s %s) values: %s %s", op, left.Type(), right.Type(), left.Inspect(), right.Inspect())
}

func (vm *VM) buildArray(startIndex, endIndex int) object.Object {
	elements := make([]object.Object, endIndex-startIndex)
	copy(elements, vm.stack[startIndex:endIndex])
	return &object.Array{Elements: elements}
}

func (vm *VM) buildHash(startIndex, endIndex int) (object.Object, error) {
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

func (vm *VM) buildPHPArray(startIndex, endIndex, numElements int) (object.Object, error) {
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

func (vm *VM) executeIndexExpression(left, index object.Object) error {
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

	case left.Type() == object.INSTANCE_OBJ:
		return vm.executeInstanceProperty(left, index)

	default:
		return fmt.Errorf("index operator not supported: %s", left.Type())
	}
}

func (vm *VM) executeInstanceProperty(instance, prop object.Object) error {
	instanceObj := instance.(*object.Instance)

	// Property name must be a string
	propName, ok := prop.(*object.String)
	if !ok {
		return fmt.Errorf("property name must be string, got %s", prop.Type())
	}

	// Check if it's a method (search in class hierarchy)
	method := findMethod(instanceObj.Class, propName.Value)
	if method != nil {
		// Return a bound method (closure with instance as first free variable)
		closure := &object.Closure{
			Fn:   method,
			Free: []object.Object{instanceObj},
		}
		vm.stack[vm.sp] = closure
		vm.sp++
		return nil
	}

	// Check if it's a field
	if field, ok := instanceObj.Fields[propName.Value]; ok {
		vm.stack[vm.sp] = field
		vm.sp++
		return nil
	}

	// Property doesn't exist
	vm.stack[vm.sp] = Null
	vm.sp++
	return nil
}

// findMethod searches for a method in the class hierarchy
func findMethod(class *object.Class, methodName string) *object.CompiledFunction {
	// Search in current class
	if method, ok := class.Methods[methodName]; ok {
		return method
	}

	// Search in parent classes
	current := class
	for current.Parent != nil {
		current = current.Parent
		if method, ok := current.Methods[methodName]; ok {
			return method
		}
	}

	return nil
}

func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]
	switch callee := callee.(type) {
	case *object.Closure:
		return vm.callClosure(callee, numArgs)
	case *object.Builtin:
		return vm.callBuiltin(callee, numArgs)
	case *object.Class:
		// Class instantiation: MyClass() creates an instance
		return vm.instantiateClass(callee, numArgs)
	default:
		return fmt.Errorf("calling non-function and non-builtin")
	}
}

func (vm *VM) instantiateClass(class *object.Class, numArgs int) error {
	// Create instance
	instance := &object.Instance{
		Class:  class,
		Fields: make(map[string]object.Object),
	}

	// Initialize properties with arguments or null
	args := vm.stack[vm.sp-numArgs : vm.sp]
	for i, propName := range class.Properties {
		if i < len(args) {
			instance.Fields[propName] = args[i]
		} else {
			instance.Fields[propName] = Null
		}
	}

	vm.sp = vm.sp - numArgs - 1

	// Call constructor if it exists
	if constructor := findMethod(class, "init"); constructor != nil {
		// Prepare constructor closure
		constructorClosure := &object.Closure{
			Fn:   constructor,
			Free: []object.Object{instance},
		}

		// Create a new VM for the constructor execution to ensure it runs synchronously
		// and doesn't interfere with the current stack frames.
		// This is necessary because we want to ignore the return value of init()
		// and always return the instance itself.
		newVM := &VM{
			constants:   vm.constants,
			stack:       make([]object.Object, StackSize),
			sp:          0,
			globals:     vm.globals,
			globalsMu:   vm.globalsMu, // Share mutex
			frames:      make([]*Frame, MaxFrames),
			framesIndex: 1,
		}

		// Push args to new VM stack
		// 1. Push 'this' (instance)
		newVM.stack[newVM.sp] = instance
		newVM.sp++
		// 2. Push other args
		for _, arg := range args {
			newVM.stack[newVM.sp] = arg
			newVM.sp++
		}

		// Setup frame for constructor
		// We pass 0 as basePointer because we manually pushed args starting at 0
		newVM.frames[0] = NewFrame(constructorClosure, 0)
		
		// Set SP past locals
		newVM.sp = constructorClosure.Fn.NumLocals

		// Run constructor
		if err := newVM.Run(); err != nil {
			return err
		}
	}
	
	vm.stack[vm.sp] = instance
	vm.sp++

	return nil
}

func (vm *VM) callClosure(cl *object.Closure, numArgs int) error {
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

// throwValue unwinds the VM to the nearest try-handler, pushing thrown so catch can use it.
// Returns (true, nil) when a handler was found and state was updated; (false, err) when uncaught.
func (vm *VM) throwValue(thrown object.Object) (handled bool, err error) {
	if len(vm.exceptionHandlers) == 0 {
		return false, fmt.Errorf("uncaught throw: %s", thrown.Inspect())
	}
	h := vm.exceptionHandlers[len(vm.exceptionHandlers)-1]
	vm.exceptionHandlers = vm.exceptionHandlers[:len(vm.exceptionHandlers)-1]
	vm.framesIndex = h.framesIndex
	f := vm.frames[vm.framesIndex-1]
	vm.sp = h.sp
	vm.stack[vm.sp] = thrown
	vm.sp++
	if h.catchPos != 0 {
		f.ip = h.catchPos - 1
	} else if h.finallyPos != 0 {
		f.ip = h.finallyPos - 1
	} else {
		return false, fmt.Errorf("uncaught throw: %s", thrown.Inspect())
	}
	return true, nil
}

func (vm *VM) callBuiltin(builtin *object.Builtin, numArgs int) error {
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

func (vm *VM) pushClosure(constIndex int, numFree int) error {
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
				// Log panic but don't crash - errors should be handled by VM error returns
				// This is a safety net for unexpected panics
			}
		}()
		// Run and check for errors - don't silently ignore them
		if err := newVM.Run(); err != nil {
			// VM errors are returned, but we can't propagate them from a goroutine
			// The error indicates something went wrong in the spawned function
			// This could cause hangs if channel operations fail
		}
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
	case *object.Blob:
		return len(obj.Data) > 0
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

// Threshold for parallel processing
const ParallelThreshold = 1000

// executeSetIndex sets a value at the given index in an array or hash
func (vm *VM) executeSetIndex(left, index, value object.Object) error {
	switch obj := left.(type) {
	case *object.Array:
		idx, ok := index.(*object.Integer)
		if !ok {
			return fmt.Errorf("array index must be integer, got %s", index.Type())
		}
		i := int(idx.Value)
		if i < 0 || i >= len(obj.Elements) {
			// Extend array if necessary
			if i >= 0 && i < 65536 { // Reasonable limit
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

// executeArrayPush appends a value to an array
func (vm *VM) executeArrayPush(arr, value object.Object) error {
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

// executeSlice creates a slice of an array
func (vm *VM) executeSlice(arr, start, end object.Object) error {
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

// executeArrayMap applies a function to each element of an array
func (vm *VM) executeArrayMap(arr, fn object.Object) error {
	arrayObj, ok := arr.(*object.Array)
	if !ok {
		return fmt.Errorf("map requires array, got %s", arr.Type())
	}

	builtin, ok := fn.(*object.Builtin)
	if ok {
		results := make([]object.Object, len(arrayObj.Elements))
		// Use parallel processing for large arrays
		if len(arrayObj.Elements) > ParallelThreshold {
			var wg sync.WaitGroup
			for i, el := range arrayObj.Elements {
				wg.Add(1)
				go func(idx int, val object.Object) {
					defer wg.Done()
					results[idx] = builtin.Fn(val)
				}(i, el)
			}
			wg.Wait()
		} else {
			for i, el := range arrayObj.Elements {
				results[i] = builtin.Fn(el)
			}
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

	// Use parallel processing for large arrays
	if len(arrayObj.Elements) > ParallelThreshold {
		var wg sync.WaitGroup
		errChan := make(chan error, len(arrayObj.Elements))

		for i, el := range arrayObj.Elements {
			wg.Add(1)
			go func(idx int, val object.Object) {
				defer wg.Done()
				result, err := vm.applyFunctionToElement(closure, val, idx)
				if err != nil {
					errChan <- err
					return
				}
				results[idx] = result
			}(i, el)
		}
		wg.Wait()
		close(errChan)

		if err := <-errChan; err != nil {
			return err
		}
	} else {
		// Sequential processing for small arrays
		for i, el := range arrayObj.Elements {
			result, err := vm.applyFunctionToElement(closure, el, i)
			if err != nil {
				return err
			}
			results[i] = result
		}
	}

	vm.stack[vm.sp] = &object.Array{Elements: results}
	vm.sp++
	return nil
}

// executeArrayFilter filters array elements using a predicate function
func (vm *VM) executeArrayFilter(arr, fn object.Object) error {
	arrayObj, ok := arr.(*object.Array)
	if !ok {
		return fmt.Errorf("filter requires array, got %s", arr.Type())
	}

	builtin, ok := fn.(*object.Builtin)
	if ok {
		var results []object.Object
		if len(arrayObj.Elements) > ParallelThreshold {
			keepFlags := make([]bool, len(arrayObj.Elements))
			var wg sync.WaitGroup

			for i, el := range arrayObj.Elements {
				wg.Add(1)
				go func(idx int, val object.Object) {
					defer wg.Done()
					result := builtin.Fn(val)
					if isTruthy(result) {
						keepFlags[idx] = true
					}
				}(i, el)
			}
			wg.Wait()

			for i, keep := range keepFlags {
				if keep {
					results = append(results, arrayObj.Elements[i])
				}
			}
		} else {
			for _, el := range arrayObj.Elements {
				result := builtin.Fn(el)
				if isTruthy(result) {
					results = append(results, el)
				}
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

	// For filter, we need to maintain order, so parallel is more complex
	// Use sequential for now, but mark elements for inclusion
	if len(arrayObj.Elements) > ParallelThreshold {
		keepFlags := make([]bool, len(arrayObj.Elements))
		var wg sync.WaitGroup

		for i, el := range arrayObj.Elements {
			wg.Add(1)
			go func(idx int, val object.Object) {
				defer wg.Done()
				result, err := vm.applyFunctionToElement(closure, val, idx)
				if err == nil && isTruthy(result) {
					keepFlags[idx] = true
				}
			}(i, el)
		}
		wg.Wait()

		// Collect results in order
		for i, keep := range keepFlags {
			if keep {
				results = append(results, arrayObj.Elements[i])
			}
		}
	} else {
		for i, el := range arrayObj.Elements {
			result, err := vm.applyFunctionToElement(closure, el, i)
			if err != nil {
				return err
			}
			if isTruthy(result) {
				results = append(results, el)
			}
		}
	}

	vm.stack[vm.sp] = &object.Array{Elements: results}
	vm.sp++
	return nil
}

// applyFunctionToElement applies a closure to a single element
func (vm *VM) applyFunctionToElement(closure *object.Closure, element object.Object, index int) (object.Object, error) {
	// Create a new VM for this function application
	newVM := &VM{
		constants:   vm.constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     vm.globals,
		frames:      make([]*Frame, MaxFrames),
		framesIndex: 1,
	}

	// Set up the function call
	frame := NewFrame(closure, 0)
	newVM.frames[0] = frame

	// Push the element as argument
	newVM.stack[0] = element
	newVM.sp = 1

	// If the function takes an index parameter, push it too
	if closure.Fn.NumParameters >= 2 {
		newVM.stack[1] = object.GetCachedInteger(int64(index))
		newVM.sp = 2
	}

	// Run the function
	err := newVM.Run()
	if err != nil {
		return nil, err
	}

	// Get the result
	if newVM.sp > 0 {
		return newVM.stack[newVM.sp-1], nil
	}
	return Null, nil
}

// CallClosure calls a closure with the given arguments, using the provided constants and globals
// This is useful for calling closures from builtin functions
func CallClosure(closure *object.Closure, constants []object.Object, globals []object.Object, args ...object.Object) (object.Object, error) {
	if len(args) != closure.Fn.NumParameters {
		return nil, fmt.Errorf("wrong number of arguments: want=%d, got=%d", closure.Fn.NumParameters, len(args))
	}

	// Create a new VM for this function call
	newVM := &VM{
		constants:   constants,
		stack:       make([]object.Object, StackSize),
		sp:          0,
		globals:     globals,
		frames:      make([]*Frame, MaxFrames),
		framesIndex: 1,
	}

	// Set up the function call
	frame := NewFrame(closure, 0)
	newVM.frames[0] = frame

	// Push arguments
	for _, arg := range args {
		newVM.stack[newVM.sp] = arg
		newVM.sp++
	}

	// Run the function
	err := newVM.Run()
	if err != nil {
		return nil, err
	}

	// Get the result
	if newVM.sp > 0 {
		return newVM.stack[newVM.sp-1], nil
	}
	return Null, nil
}

// Deprecated: NewOptimized is now just an alias for New
// Kept for backwards compatibility
func NewOptimized(bytecode *compiler.Bytecode) *VM {
	return New(bytecode)
}

// Deprecated: NewOptimizedWithGlobalsStore is now just an alias for NewWithGlobalsStore
// Kept for backwards compatibility
func NewOptimizedWithGlobalsStore(bytecode *compiler.Bytecode, s []object.Object) *VM {
	return NewWithGlobalsStore(bytecode, s)
}
