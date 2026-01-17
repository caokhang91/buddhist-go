package object

import (
	"fmt"
	"strings"
)

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
