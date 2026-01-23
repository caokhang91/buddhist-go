package code

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// Opcode represents a single byte instruction
type Opcode byte

// Instructions is a sequence of bytecode instructions
type Instructions []byte

// Opcodes
const (
	OpConstant           Opcode = iota // Push constant onto stack
	OpAdd                              // Add two values
	OpSub                              // Subtract two values
	OpMul                              // Multiply two values
	OpDiv                              // Divide two values
	OpMod                              // Modulo two values
	OpPop                              // Pop top of stack
	OpTrue                             // Push true onto stack
	OpFalse                            // Push false onto stack
	OpNull                             // Push null onto stack
	OpEqual                            // Compare equality
	OpNotEqual                         // Compare inequality
	OpGreaterThan                      // Greater than comparison
	OpGreaterThanOrEqual               // Greater than or equal comparison
	OpLessThan                         // Less than comparison
	OpLessThanOrEqual                  // Less than or equal comparison
	OpAnd                              // Logical AND
	OpOr                               // Logical OR
	OpMinus                            // Negate value
	OpBang                             // Logical NOT
	OpJump                             // Unconditional jump
	OpJumpNotTruthy                    // Jump if not truthy
	OpGetGlobal                        // Get global variable
	OpSetGlobal                        // Set global variable
	OpGetLocal                         // Get local variable
	OpSetLocal                         // Set local variable
	OpArray                            // Create array
	OpHash                             // Create hash
	OpIndex                            // Index operation
	OpCall                             // Call function
	OpReturnValue                      // Return with value
	OpReturn                           // Return without value
	OpGetBuiltin                       // Get builtin function
	OpClosure                          // Create closure
	OpGetFree                          // Get free variable
	OpCurrentClosure                   // Get current closure
	OpSpawn                            // Spawn goroutine
	OpChannel                          // Create channel
	OpChannelBuffered                  // Create buffered channel
	OpSend                             // Send to channel
	OpReceive                          // Receive from channel
	OpCloseChannel                     // Close channel
	OpBreak                            // Break from loop
	OpContinue                         // Continue loop
	// Class opcodes
	OpClass       // Push class constant
	OpInstantiate // Instantiate class
	OpGetMethod   // Get method
	OpCallMethod  // Call method
	OpGetProperty // Get instance property
	OpSetProperty // Set instance property
	OpThis        // Push 'this' (current instance)
	OpSuper       // Push 'super' (parent class instance)
	OpInherit     // Inherit from parent class
	// Error handling opcodes
	OpTry     // Try block
	OpThrow   // Throw statement
	OpFinally // Finally block
	// Array opcodes (PHP-style operations)
	OpSetIndex    // arr[key] = val - Set element at index
	OpArrayPush   // arr[] = val - Push with auto-increment
	OpSlice       // arr[start:end] - Slice array
	OpArrayMap    // Map function over array
	OpArrayFilter // Filter array with function
	OpPHPArray    // Create PHP-style array (ordered map)
)

// Definition describes an opcode
type Definition struct {
	Name          string
	OperandWidths []int // Width of each operand in bytes
}

var definitions = map[Opcode]*Definition{
	OpConstant:           {"OpConstant", []int{2}},
	OpAdd:                {"OpAdd", []int{}},
	OpSub:                {"OpSub", []int{}},
	OpMul:                {"OpMul", []int{}},
	OpDiv:                {"OpDiv", []int{}},
	OpMod:                {"OpMod", []int{}},
	OpPop:                {"OpPop", []int{}},
	OpTrue:               {"OpTrue", []int{}},
	OpFalse:              {"OpFalse", []int{}},
	OpNull:               {"OpNull", []int{}},
	OpEqual:              {"OpEqual", []int{}},
	OpNotEqual:           {"OpNotEqual", []int{}},
	OpGreaterThan:        {"OpGreaterThan", []int{}},
	OpGreaterThanOrEqual: {"OpGreaterThanOrEqual", []int{}},
	OpLessThan:           {"OpLessThan", []int{}},
	OpLessThanOrEqual:    {"OpLessThanOrEqual", []int{}},
	OpAnd:                {"OpAnd", []int{}},
	OpOr:                 {"OpOr", []int{}},
	OpMinus:              {"OpMinus", []int{}},
	OpBang:               {"OpBang", []int{}},
	OpJump:               {"OpJump", []int{2}},
	OpJumpNotTruthy:      {"OpJumpNotTruthy", []int{2}},
	OpGetGlobal:          {"OpGetGlobal", []int{2}},
	OpSetGlobal:          {"OpSetGlobal", []int{2}},
	OpGetLocal:           {"OpGetLocal", []int{1}},
	OpSetLocal:           {"OpSetLocal", []int{1}},
	OpArray:              {"OpArray", []int{2}},
	OpHash:               {"OpHash", []int{2}},
	OpIndex:              {"OpIndex", []int{}},
	OpCall:               {"OpCall", []int{1}},
	OpReturnValue:        {"OpReturnValue", []int{}},
	OpReturn:             {"OpReturn", []int{}},
	OpGetBuiltin:         {"OpGetBuiltin", []int{1}},
	OpClosure:            {"OpClosure", []int{2, 1}},
	OpGetFree:            {"OpGetFree", []int{1}},
	OpCurrentClosure:     {"OpCurrentClosure", []int{}},
	OpSpawn:              {"OpSpawn", []int{}},
	OpChannel:            {"OpChannel", []int{}},
	OpChannelBuffered:    {"OpChannelBuffered", []int{}}, // Buffer size from stack
	OpSend:               {"OpSend", []int{}},
	OpReceive:            {"OpReceive", []int{}},
	OpCloseChannel:       {"OpCloseChannel", []int{}},
	OpBreak:              {"OpBreak", []int{}},
	OpContinue:           {"OpContinue", []int{}},
	// Class opcodes
	OpClass:       {"OpClass", []int{2}},
	OpInstantiate: {"OpInstantiate", []int{1}},
	OpGetMethod:   {"OpGetMethod", []int{2}},
	OpCallMethod:  {"OpCallMethod", []int{1}},
	OpGetProperty: {"OpGetProperty", []int{}},
	OpSetProperty: {"OpSetProperty", []int{}},
	OpThis:        {"OpThis", []int{}},
	OpSuper:       {"OpSuper", []int{}},
	OpInherit:     {"OpInherit", []int{}},
	// Error handling opcodes
	OpTry:     {"OpTry", []int{2, 2}},
	OpThrow:   {"OpThrow", []int{}},
	OpFinally: {"OpFinally", []int{2}},
	// Array opcodes
	OpSetIndex:    {"OpSetIndex", []int{}},
	OpArrayPush:   {"OpArrayPush", []int{}},
	OpSlice:       {"OpSlice", []int{}},
	OpArrayMap:    {"OpArrayMap", []int{}},
	OpArrayFilter: {"OpArrayFilter", []int{}},
	OpPHPArray:    {"OpPHPArray", []int{2}},
}

// Lookup returns the definition for an opcode
func Lookup(op byte) (*Definition, error) {
	def, ok := definitions[Opcode(op)]
	if !ok {
		return nil, fmt.Errorf("opcode %d undefined", op)
	}
	return def, nil
}

// Make creates a bytecode instruction
func Make(op Opcode, operands ...int) []byte {
	def, ok := definitions[op]
	if !ok {
		return []byte{}
	}

	instructionLen := 1
	for _, w := range def.OperandWidths {
		instructionLen += w
	}

	instruction := make([]byte, instructionLen)
	instruction[0] = byte(op)

	offset := 1
	for i, o := range operands {
		width := def.OperandWidths[i]
		switch width {
		case 2:
			binary.BigEndian.PutUint16(instruction[offset:], uint16(o))
		case 1:
			instruction[offset] = byte(o)
		}
		offset += width
	}

	return instruction
}

// ReadOperands reads operands from instructions
func ReadOperands(def *Definition, ins Instructions) ([]int, int) {
	operands := make([]int, len(def.OperandWidths))
	offset := 0

	for i, width := range def.OperandWidths {
		switch width {
		case 2:
			operands[i] = int(ReadUint16(ins[offset:]))
		case 1:
			operands[i] = int(ReadUint8(ins[offset:]))
		}
		offset += width
	}

	return operands, offset
}

// ReadUint16 reads a uint16 from instructions
func ReadUint16(ins Instructions) uint16 {
	return binary.BigEndian.Uint16(ins)
}

// ReadUint8 reads a uint8 from instructions
func ReadUint8(ins Instructions) uint8 {
	return uint8(ins[0])
}

// String returns a string representation of the instructions
func (ins Instructions) String() string {
	var out bytes.Buffer

	i := 0
	for i < len(ins) {
		def, err := Lookup(ins[i])
		if err != nil {
			fmt.Fprintf(&out, "ERROR: %s\n", err)
			continue
		}

		operands, read := ReadOperands(def, ins[i+1:])

		fmt.Fprintf(&out, "%04d %s\n", i, ins.fmtInstruction(def, operands))

		i += 1 + read
	}

	return out.String()
}

func (ins Instructions) fmtInstruction(def *Definition, operands []int) string {
	operandCount := len(def.OperandWidths)

	if len(operands) != operandCount {
		return fmt.Sprintf("ERROR: operand len %d does not match defined %d\n",
			len(operands), operandCount)
	}

	switch operandCount {
	case 0:
		return def.Name
	case 1:
		return fmt.Sprintf("%s %d", def.Name, operands[0])
	case 2:
		return fmt.Sprintf("%s %d %d", def.Name, operands[0], operands[1])
	}

	return fmt.Sprintf("ERROR: unhandled operandCount for %s\n", def.Name)
}
