package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"

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
	BLOB_OBJ         ObjectType = "BLOB"
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

// Copy creates a shallow copy of the array
func (a *Array) Copy() *Array {
	newElements := make([]Object, len(a.Elements))
	copy(newElements, a.Elements)
	return &Array{Elements: newElements}
}

// Slice returns a slice of the array from start to end (exclusive)
func (a *Array) Slice(start, end int) *Array {
	length := len(a.Elements)
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start >= end {
		return &Array{Elements: []Object{}}
	}
	newElements := make([]Object, end-start)
	copy(newElements, a.Elements[start:end])
	return &Array{Elements: newElements}
}

// MapEntry represents a key-value pair in a PHPArray
type MapEntry struct {
	Key   Object
	Value Object
}

// PHPArray represents a PHP-style array (ordered hash map)
// It maintains insertion order while providing O(1) key lookup
type PHPArray struct {
	Entries    []MapEntry
	Indices    map[interface{}]int
	NextIntKey int64 // Auto-increment key for push operations
}

// NewPHPArray creates a new PHP-style array
func NewPHPArray() *PHPArray {
	return &PHPArray{
		Entries:    []MapEntry{},
		Indices:    make(map[interface{}]int),
		NextIntKey: 0,
	}
}

func (p *PHPArray) Type() ObjectType { return ARRAY_OBJ }
func (p *PHPArray) Inspect() string {
	var out bytes.Buffer
	pairs := []string{}
	for _, entry := range p.Entries {
		pairs = append(pairs, fmt.Sprintf("%s => %s", entry.Key.Inspect(), entry.Value.Inspect()))
	}
	out.WriteString("[")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("]")
	return out.String()
}

// getHashKey returns a hashable key for map lookup
func (p *PHPArray) getHashKey(key Object) interface{} {
	switch k := key.(type) {
	case *Integer:
		return k.Value
	case *String:
		return k.Value
	case *Boolean:
		if k.Value {
			return int64(1)
		}
		return int64(0)
	default:
		return key.Inspect()
	}
}

// Set sets a value at the given key
func (p *PHPArray) Set(key Object, value Object) {
	hashKey := p.getHashKey(key)

	// Update NextIntKey if key is an integer
	if intKey, ok := key.(*Integer); ok {
		if intKey.Value >= p.NextIntKey {
			p.NextIntKey = intKey.Value + 1
		}
	}

	if idx, exists := p.Indices[hashKey]; exists {
		// Update existing entry
		p.Entries[idx].Value = value
	} else {
		// Add new entry
		p.Entries = append(p.Entries, MapEntry{Key: key, Value: value})
		p.Indices[hashKey] = len(p.Entries) - 1
	}
}

// Get retrieves a value by key
func (p *PHPArray) Get(key Object) (Object, bool) {
	hashKey := p.getHashKey(key)
	if idx, exists := p.Indices[hashKey]; exists {
		return p.Entries[idx].Value, true
	}
	return nil, false
}

// Push appends a value with auto-incrementing integer key
func (p *PHPArray) Push(value Object) {
	key := &Integer{Value: p.NextIntKey}
	p.NextIntKey++
	p.Entries = append(p.Entries, MapEntry{Key: key, Value: value})
	p.Indices[key.Value] = len(p.Entries) - 1
}

// Length returns the number of entries
func (p *PHPArray) Length() int {
	return len(p.Entries)
}

// ToArray converts PHPArray to regular Array (values only)
func (p *PHPArray) ToArray() *Array {
	elements := make([]Object, len(p.Entries))
	for i, entry := range p.Entries {
		elements[i] = entry.Value
	}
	return &Array{Elements: elements}
}

// Keys returns all keys as an array
func (p *PHPArray) Keys() *Array {
	keys := make([]Object, len(p.Entries))
	for i, entry := range p.Entries {
		keys[i] = entry.Key
	}
	return &Array{Elements: keys}
}

// Values returns all values as an array
func (p *PHPArray) Values() *Array {
	values := make([]Object, len(p.Entries))
	for i, entry := range p.Entries {
		values[i] = entry.Value
	}
	return &Array{Elements: values}
}

// Copy creates a deep copy of PHPArray
func (p *PHPArray) Copy() *PHPArray {
	newArr := NewPHPArray()
	newArr.NextIntKey = p.NextIntKey
	for _, entry := range p.Entries {
		newArr.Set(entry.Key, entry.Value)
	}
	return newArr
}

// Delete removes an entry by key
func (p *PHPArray) Delete(key Object) bool {
	hashKey := p.getHashKey(key)
	idx, exists := p.Indices[hashKey]
	if !exists {
		return false
	}

	// Remove from entries
	p.Entries = append(p.Entries[:idx], p.Entries[idx+1:]...)

	// Rebuild indices
	delete(p.Indices, hashKey)
	for i := idx; i < len(p.Entries); i++ {
		hk := p.getHashKey(p.Entries[i].Key)
		p.Indices[hk] = i
	}

	return true
}

// WorkerFunc is a function type for parallel operations
type WorkerFunc func(Object) Object

// ParallelMap applies a worker function to all values in parallel
// Uses Go's concurrency for large arrays (threshold: 1000 elements)
func (p *PHPArray) ParallelMap(worker WorkerFunc) *PHPArray {
	newArr := NewPHPArray()
	length := len(p.Entries)

	if length == 0 {
		return newArr
	}

	// Parallel processing threshold
	const threshold = 1000

	if length > threshold {
		// Parallel processing for large arrays
		type result struct {
			key   Object
			value Object
			order int
		}
		results := make(chan result, length)
		var wg sync.WaitGroup

		for i, entry := range p.Entries {
			wg.Add(1)
			go func(idx int, k, v Object) {
				defer wg.Done()
				results <- result{
					key:   k,
					value: worker(v),
					order: idx,
				}
			}(i, entry.Key, entry.Value)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect results maintaining order
		orderedResults := make([]result, length)
		for res := range results {
			orderedResults[res.order] = res
		}

		// Build new array in original order
		for _, res := range orderedResults {
			newArr.Set(res.key, res.value)
		}
	} else {
		// Sequential processing for small arrays
		for _, entry := range p.Entries {
			newArr.Set(entry.Key, worker(entry.Value))
		}
	}

	return newArr
}

// FilterFunc is a function type for filter operations
type FilterFunc func(Object) bool

// ParallelFilter filters values in parallel using a predicate
func (p *PHPArray) ParallelFilter(predicate FilterFunc) *PHPArray {
	newArr := NewPHPArray()
	length := len(p.Entries)

	if length == 0 {
		return newArr
	}

	const threshold = 1000

	if length > threshold {
		// Parallel processing
		type filterResult struct {
			entry MapEntry
			keep  bool
			order int
		}
		results := make(chan filterResult, length)
		var wg sync.WaitGroup

		for i, entry := range p.Entries {
			wg.Add(1)
			go func(idx int, e MapEntry) {
				defer wg.Done()
				results <- filterResult{
					entry: e,
					keep:  predicate(e.Value),
					order: idx,
				}
			}(i, entry)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect and sort by order
		orderedResults := make([]filterResult, length)
		for res := range results {
			orderedResults[res.order] = res
		}

		// Build filtered array in order
		for _, res := range orderedResults {
			if res.keep {
				newArr.Set(res.entry.Key, res.entry.Value)
			}
		}
	} else {
		// Sequential processing
		for _, entry := range p.Entries {
			if predicate(entry.Value) {
				newArr.Set(entry.Key, entry.Value)
			}
		}
	}

	return newArr
}

// Reduce reduces the PHPArray to a single value
func (p *PHPArray) Reduce(fn func(accumulator, current Object) Object, initial Object) Object {
	acc := initial
	for _, entry := range p.Entries {
		acc = fn(acc, entry.Value)
	}
	return acc
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
	Parent     *Class   // Parent class for inheritance
	ParentName string   // Parent class name (for deferred resolution)
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
