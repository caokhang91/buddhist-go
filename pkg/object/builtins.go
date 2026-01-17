package object

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
)

// ParallelThreshold defines when to use parallel processing
const ParallelThreshold = 1000

// CompiledFunction represents a compiled function
type CompiledFunction struct {
	Instructions  []byte
	NumLocals     int
	NumParameters int
}

func (cf *CompiledFunction) Type() ObjectType { return FUNCTION_OBJ }
func (cf *CompiledFunction) Inspect() string {
	return fmt.Sprintf("CompiledFunction[%p]", cf)
}

// Closure wraps a compiled function with free variables
type Closure struct {
	Fn   *CompiledFunction
	Free []Object
}

func (c *Closure) Type() ObjectType { return FUNCTION_OBJ }
func (c *Closure) Inspect() string {
	return fmt.Sprintf("Closure[%p]", c)
}

// BuiltinDef represents a builtin function definition
type BuiltinDef struct {
	Name string
	Fn   BuiltinFunction
}

// Builtins is the list of builtin functions
var Builtins = []BuiltinDef{
	{
		Name: "len",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *String:
				return &Integer{Value: int64(len(arg.Value))}
			case *Array:
				return &Integer{Value: int64(len(arg.Elements))}
			case *Blob:
				return &Integer{Value: int64(len(arg.Data))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	{
		Name: "print",
		Fn: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Print(arg.Inspect())
			}
			return &Null{}
		},
	},
	{
		Name: "println",
		Fn: func(args ...Object) Object {
			for _, arg := range args {
				fmt.Println(arg.Inspect())
			}
			return &Null{}
		},
	},
	{
		Name: "first",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `first` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			if len(arr.Elements) > 0 {
				return arr.Elements[0]
			}
			return &Null{}
		},
	},
	{
		Name: "last",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `last` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			length := len(arr.Elements)
			if length > 0 {
				return arr.Elements[length-1]
			}
			return &Null{}
		},
	},
	{
		Name: "rest",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `rest` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			length := len(arr.Elements)
			if length > 0 {
				newElements := make([]Object, length-1)
				copy(newElements, arr.Elements[1:length])
				return &Array{Elements: newElements}
			}
			return &Null{}
		},
	},
	{
		Name: "push",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `push` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			length := len(arr.Elements)
			newElements := make([]Object, length+1)
			copy(newElements, arr.Elements)
			newElements[length] = args[1]
			return &Array{Elements: newElements}
		},
	},
	{
		Name: "type",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			return &String{Value: string(args[0].Type())}
		},
	},
	{
		Name: "str",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			return &String{Value: args[0].Inspect()}
		},
	},
	{
		Name: "int",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *Integer:
				return arg
			case *Float:
				return &Integer{Value: int64(arg.Value)}
			case *String:
				var i int64
				_, err := fmt.Sscanf(arg.Value, "%d", &i)
				if err != nil {
					return newError("cannot convert %q to integer", arg.Value)
				}
				return &Integer{Value: i}
			case *Boolean:
				if arg.Value {
					return &Integer{Value: 1}
				}
				return &Integer{Value: 0}
			default:
				return newError("cannot convert %s to integer", args[0].Type())
			}
		},
	},
	{
		Name: "float",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *Integer:
				return &Float{Value: float64(arg.Value)}
			case *Float:
				return arg
			case *String:
				var f float64
				_, err := fmt.Sscanf(arg.Value, "%f", &f)
				if err != nil {
					return newError("cannot convert %q to float", arg.Value)
				}
				return &Float{Value: f}
			default:
				return newError("cannot convert %s to float", args[0].Type())
			}
		},
	},
	{
		Name: "split",
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=1 or 2", len(args))
			}
			if args[0].Type() != STRING_OBJ {
				return newError("first argument to `split` must be STRING, got %s", args[0].Type())
			}
			str := args[0].(*String).Value
			sep := " "
			if len(args) == 2 {
				if args[1].Type() != STRING_OBJ {
					return newError("second argument to `split` must be STRING, got %s", args[1].Type())
				}
				sep = args[1].(*String).Value
			}
			parts := strings.Split(str, sep)
			elements := make([]Object, len(parts))
			for i, p := range parts {
				elements[i] = &String{Value: p}
			}
			return &Array{Elements: elements}
		},
	},
	{
		Name: "join",
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 2 {
				return newError("wrong number of arguments. got=%d, want=1 or 2", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("first argument to `join` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			sep := ""
			if len(args) == 2 {
				if args[1].Type() != STRING_OBJ {
					return newError("second argument to `join` must be STRING, got %s", args[1].Type())
				}
				sep = args[1].(*String).Value
			}
			parts := make([]string, len(arr.Elements))
			for i, el := range arr.Elements {
				parts[i] = el.Inspect()
			}
			return &String{Value: strings.Join(parts, sep)}
		},
	},
	{
		Name: "slice",
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 3 {
				return newError("wrong number of arguments. got=%d, want=1 to 3", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("first argument to `slice` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			length := len(arr.Elements)
			start := 0
			end := length

			if len(args) >= 2 {
				if args[1].Type() != INTEGER_OBJ {
					return newError("second argument to `slice` must be INTEGER, got %s", args[1].Type())
				}
				start = int(args[1].(*Integer).Value)
			}
			if len(args) >= 3 {
				if args[2].Type() != INTEGER_OBJ {
					return newError("third argument to `slice` must be INTEGER, got %s", args[2].Type())
				}
				end = int(args[2].(*Integer).Value)
			}

			// Handle negative indices
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
			copy(newElements, arr.Elements[start:end])
			return &Array{Elements: newElements}
		},
	},
	{
		Name: "range",
		Fn: func(args ...Object) Object {
			if len(args) < 1 || len(args) > 3 {
				return newError("wrong number of arguments. got=%d, want=1 to 3", len(args))
			}

			var start, end, step int64 = 0, 0, 1

			if len(args) == 1 {
				if args[0].Type() != INTEGER_OBJ {
					return newError("argument to `range` must be INTEGER, got %s", args[0].Type())
				}
				end = args[0].(*Integer).Value
			} else if len(args) >= 2 {
				if args[0].Type() != INTEGER_OBJ || args[1].Type() != INTEGER_OBJ {
					return newError("arguments to `range` must be INTEGER")
				}
				start = args[0].(*Integer).Value
				end = args[1].(*Integer).Value
			}
			if len(args) == 3 {
				if args[2].Type() != INTEGER_OBJ {
					return newError("third argument to `range` must be INTEGER, got %s", args[2].Type())
				}
				step = args[2].(*Integer).Value
				if step == 0 {
					return newError("step cannot be zero")
				}
			}

			var elements []Object
			if step > 0 {
				for i := start; i < end; i += step {
					elements = append(elements, &Integer{Value: i})
				}
			} else {
				for i := start; i > end; i += step {
					elements = append(elements, &Integer{Value: i})
				}
			}

			return &Array{Elements: elements}
		},
	},
	{
		Name: "map",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("first argument to `map` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)

			// For builtin map, we need to handle this differently
			// The second argument should be a builtin function
			builtin, ok := args[1].(*Builtin)
			if !ok {
				return newError("second argument to `map` must be a builtin function, got %s", args[1].Type())
			}

			results := make([]Object, len(arr.Elements))

			// Use parallel processing for large arrays
			if len(arr.Elements) > ParallelThreshold {
				var wg sync.WaitGroup
				for i, el := range arr.Elements {
					wg.Add(1)
					go func(idx int, val Object) {
						defer wg.Done()
						results[idx] = builtin.Fn(val)
					}(i, el)
				}
				wg.Wait()
			} else {
				for i, el := range arr.Elements {
					results[i] = builtin.Fn(el)
				}
			}

			return &Array{Elements: results}
		},
	},
	{
		Name: "filter",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("first argument to `filter` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)

			builtin, ok := args[1].(*Builtin)
			if !ok {
				return newError("second argument to `filter` must be a builtin function, got %s", args[1].Type())
			}

			var results []Object

			if len(arr.Elements) > ParallelThreshold {
				keepFlags := make([]bool, len(arr.Elements))
				var wg sync.WaitGroup

				for i, el := range arr.Elements {
					wg.Add(1)
					go func(idx int, val Object) {
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
						results = append(results, arr.Elements[i])
					}
				}
			} else {
				for _, el := range arr.Elements {
					result := builtin.Fn(el)
					if isTruthy(result) {
						results = append(results, el)
					}
				}
			}

			return &Array{Elements: results}
		},
	},
	{
		Name: "reduce",
		Fn: func(args ...Object) Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments. got=%d, want=2 or 3", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("first argument to `reduce` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)

			if len(arr.Elements) == 0 {
				if len(args) == 3 {
					return args[2]
				}
				return &Null{}
			}

			builtin, ok := args[1].(*Builtin)
			if !ok {
				return newError("second argument to `reduce` must be a builtin function, got %s", args[1].Type())
			}

			var accumulator Object
			startIdx := 0

			if len(args) == 3 {
				accumulator = args[2]
			} else {
				accumulator = arr.Elements[0]
				startIdx = 1
			}

			for i := startIdx; i < len(arr.Elements); i++ {
				accumulator = builtin.Fn(accumulator, arr.Elements[i])
				if _, ok := accumulator.(*Error); ok {
					return accumulator
				}
			}

			return accumulator
		},
	},
	{
		Name: "reverse",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `reverse` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			length := len(arr.Elements)
			newElements := make([]Object, length)
			for i, el := range arr.Elements {
				newElements[length-1-i] = el
			}
			return &Array{Elements: newElements}
		},
	},
	{
		Name: "concat",
		Fn: func(args ...Object) Object {
			if len(args) < 2 {
				return newError("wrong number of arguments. got=%d, want>=2", len(args))
			}
			var elements []Object
			for _, arg := range args {
				if arg.Type() != ARRAY_OBJ {
					return newError("arguments to `concat` must be ARRAY, got %s", arg.Type())
				}
				arr := arg.(*Array)
				elements = append(elements, arr.Elements...)
			}
			return &Array{Elements: elements}
		},
	},
	{
		Name: "contains",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("first argument to `contains` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			target := args[1]

			for _, el := range arr.Elements {
				if objectsEqual(el, target) {
					return &Boolean{Value: true}
				}
			}
			return &Boolean{Value: false}
		},
	},
	{
		Name: "indexOf",
		Fn: func(args ...Object) Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("first argument to `indexOf` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			target := args[1]

			for i, el := range arr.Elements {
				if objectsEqual(el, target) {
					return &Integer{Value: int64(i)}
				}
			}
			return &Integer{Value: -1}
		},
	},
	{
		Name: "unique",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `unique` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			seen := make(map[string]bool)
			var unique []Object

			for _, el := range arr.Elements {
				key := el.Inspect()
				if !seen[key] {
					seen[key] = true
					unique = append(unique, el)
				}
			}
			return &Array{Elements: unique}
		},
	},
	{
		Name: "flatten",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `flatten` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			return &Array{Elements: flattenArray(arr)}
		},
	},
	{
		Name: "sum",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `sum` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)

			var intSum int64
			var floatSum float64
			hasFloat := false

			for _, el := range arr.Elements {
				switch v := el.(type) {
				case *Integer:
					intSum += v.Value
				case *Float:
					hasFloat = true
					floatSum += v.Value
				default:
					return newError("sum requires array of numbers, got %s", el.Type())
				}
			}

			if hasFloat {
				return &Float{Value: floatSum + float64(intSum)}
			}
			return &Integer{Value: intSum}
		},
	},
	{
		Name: "min",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `min` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			if len(arr.Elements) == 0 {
				return &Null{}
			}

			min := arr.Elements[0]
			for _, el := range arr.Elements[1:] {
				if compareNumbers(el, min) < 0 {
					min = el
				}
			}
			return min
		},
	},
	{
		Name: "max",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `max` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			if len(arr.Elements) == 0 {
				return &Null{}
			}

			max := arr.Elements[0]
			for _, el := range arr.Elements[1:] {
				if compareNumbers(el, max) > 0 {
					max = el
				}
			}
			return max
		},
	},
	{
		Name: "avg",
		Fn: func(args ...Object) Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if args[0].Type() != ARRAY_OBJ {
				return newError("argument to `avg` must be ARRAY, got %s", args[0].Type())
			}
			arr := args[0].(*Array)
			if len(arr.Elements) == 0 {
				return &Null{}
			}

			var sum float64
			for _, el := range arr.Elements {
				switch v := el.(type) {
				case *Integer:
					sum += float64(v.Value)
				case *Float:
					sum += v.Value
				default:
					return newError("avg requires array of numbers, got %s", el.Type())
				}
			}

			return &Float{Value: sum / float64(len(arr.Elements))}
		},
	},
	{
		Name: "blob_new",
		Fn:   blobNewBuiltin,
	},
	{
		Name: "blob_from_string",
		Fn:   blobFromStringBuiltin,
	},
	{
		Name: "blob_from_file",
		Fn:   blobFromFileBuiltin,
	},
	{
		Name: "blob_write_file",
		Fn:   blobWriteFileBuiltin,
	},
	{
		Name: "blob_slice",
		Fn:   blobSliceBuiltin,
	},
	{
		Name: "blob_read_int",
		Fn:   blobReadIntBuiltin,
	},
	{
		Name: "blob_write_int",
		Fn:   blobWriteIntBuiltin,
	},
	{
		Name: "blob_read_float",
		Fn:   blobReadFloatBuiltin,
	},
	{
		Name: "blob_write_float",
		Fn:   blobWriteFloatBuiltin,
	},
	{
		Name: "blob_mmap",
		Fn:   blobMmapBuiltin,
	},
	{
		Name: "blob_unmap",
		Fn:   blobUnmapBuiltin,
	},
	{
		Name: "blob_release",
		Fn:   blobReleaseBuiltin,
	},
	{
		Name: "http_request",
		Fn:   httpRequestBuiltin,
	},
	{
		Name: "curl",
		Fn:   curlBuiltin,
	},
}

// GetBuiltinByName returns a builtin function by name
func GetBuiltinByName(name string) *Builtin {
	for _, def := range Builtins {
		if def.Name == name {
			return &Builtin{Name: def.Name, Fn: def.Fn}
		}
	}
	return nil
}

func newError(format string, a ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, a...)}
}

// isTruthy determines if an object is truthy
func isTruthy(obj Object) bool {
	switch o := obj.(type) {
	case *Boolean:
		return o.Value
	case *Null:
		return false
	case *Integer:
		return o.Value != 0
	case *String:
		return len(o.Value) > 0
	case *Array:
		return len(o.Elements) > 0
	case *Blob:
		return len(o.Data) > 0
	default:
		return true
	}
}

// objectsEqual checks if two objects are equal
func objectsEqual(a, b Object) bool {
	if a.Type() != b.Type() {
		return false
	}
	switch aVal := a.(type) {
	case *Integer:
		return aVal.Value == b.(*Integer).Value
	case *Float:
		return aVal.Value == b.(*Float).Value
	case *String:
		return aVal.Value == b.(*String).Value
	case *Boolean:
		return aVal.Value == b.(*Boolean).Value
	case *Null:
		return true
	case *Blob:
		return bytes.Equal(aVal.Data, b.(*Blob).Data)
	default:
		return a.Inspect() == b.Inspect()
	}
}

// flattenArray recursively flattens nested arrays
func flattenArray(arr *Array) []Object {
	var result []Object
	for _, el := range arr.Elements {
		if nested, ok := el.(*Array); ok {
			result = append(result, flattenArray(nested)...)
		} else {
			result = append(result, el)
		}
	}
	return result
}

// compareNumbers compares two numeric objects
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func compareNumbers(a, b Object) int {
	var aVal, bVal float64

	switch v := a.(type) {
	case *Integer:
		aVal = float64(v.Value)
	case *Float:
		aVal = v.Value
	default:
		return 0
	}

	switch v := b.(type) {
	case *Integer:
		bVal = float64(v.Value)
	case *Float:
		bVal = v.Value
	default:
		return 0
	}

	if aVal < bVal {
		return -1
	} else if aVal > bVal {
		return 1
	}
	return 0
}
