package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/caokhang91/buddhist-go/pkg/ast"
)

// ObjectType represents the type of an object
type ObjectType string

// Object types
const (
	INTEGER_OBJ      ObjectType = "INTEGER"
	FLOAT_OBJ        ObjectType = "FLOAT"
	BOOLEAN_OBJ      ObjectType = "BOOLEAN"
	NULL_OBJ         ObjectType = "NULL"
	STRING_OBJ       ObjectType = "STRING"
	RETURN_VALUE_OBJ ObjectType = "RETURN_VALUE"
	ERROR_OBJ        ObjectType = "ERROR"
	FUNCTION_OBJ     ObjectType = "FUNCTION"
	BUILTIN_OBJ      ObjectType = "BUILTIN"
	ARRAY_OBJ        ObjectType = "ARRAY"
	HASH_OBJ         ObjectType = "HASH"
	CHANNEL_OBJ      ObjectType = "CHANNEL"
	BREAK_OBJ        ObjectType = "BREAK"
	CONTINUE_OBJ     ObjectType = "CONTINUE"
	CLASS_OBJ        ObjectType = "CLASS"
	INSTANCE_OBJ     ObjectType = "INSTANCE"
)

// Object interface represents all objects in the language
type Object interface {
	Type() ObjectType
	Inspect() string
}

// Hashable interface for objects that can be used as hash keys
type Hashable interface {
	HashKey() HashKey
}

// HashKey represents a key for the hash map
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Integer represents an integer value
type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

// Float represents a floating-point value
type Float struct {
	Value float64
}

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return fmt.Sprintf("%g", f.Value) }

// Boolean represents a boolean value
type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}
	return HashKey{Type: b.Type(), Value: value}
}

// Null represents a null value
type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "null" }

// String represents a string value
type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// ReturnValue wraps a value for return statements
type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

// Error represents an error
type Error struct {
	Message string
	Line    int
	Column  int
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string {
	if e.Line > 0 {
		return fmt.Sprintf("ERROR at line %d: %s", e.Line, e.Message)
	}
	return "ERROR: " + e.Message
}

// Function represents a function
type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
	Name       string
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer
	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}
	out.WriteString("fn")
	if f.Name != "" {
		out.WriteString(" " + f.Name)
	}
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")
	return out.String()
}

// BuiltinFunction represents a builtin function type
type BuiltinFunction func(args ...Object) Object

// Builtin represents a builtin function
type Builtin struct {
	Name string
	Fn   BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string  { return "builtin function: " + b.Name }

// Array represents an array
type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var out bytes.Buffer
	elements := []string{}
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}
	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}

// HashPair represents a key-value pair in a hash
type HashPair struct {
	Key   Object
	Value Object
}

// Hash represents a hash map
type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Type() ObjectType { return HASH_OBJ }
func (h *Hash) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s",
			pair.Key.Inspect(), pair.Value.Inspect()))
	}
	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// Channel represents a channel for concurrency (maps to Go channel)
type Channel struct {
	Chan chan Object
}

func (c *Channel) Type() ObjectType { return CHANNEL_OBJ }
func (c *Channel) Inspect() string  { return "channel" }

// Break represents a break signal
type Break struct{}

func (b *Break) Type() ObjectType { return BREAK_OBJ }
func (b *Break) Inspect() string  { return "break" }

// Continue represents a continue signal
type Continue struct{}

func (c *Continue) Type() ObjectType { return CONTINUE_OBJ }
func (c *Continue) Inspect() string  { return "continue" }

// Class represents a class definition
// Note: CompiledFunction is defined in builtins.go
type Class struct {
	Name       string
	Methods    map[string]*CompiledFunction
	Properties []string // Property names for initialization
}

func (c *Class) Type() ObjectType { return CLASS_OBJ }
func (c *Class) Inspect() string {
	return fmt.Sprintf("class %s", c.Name)
}

// Instance represents an instance of a class
type Instance struct {
	Class  *Class
	Fields map[string]Object
}

func (i *Instance) Type() ObjectType { return INSTANCE_OBJ }
func (i *Instance) Inspect() string {
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("%s instance", i.Class.Name))
	return out.String()
}
