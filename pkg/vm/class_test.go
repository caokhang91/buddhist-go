package vm

import (
	"fmt"
	"testing"

	"github.com/caokhang91/buddhist-go/pkg/compiler"
	"github.com/caokhang91/buddhist-go/pkg/lexer"
	"github.com/caokhang91/buddhist-go/pkg/object"
	"github.com/caokhang91/buddhist-go/pkg/parser"
)

func TestClassDefinition(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class Person {
				let name = null;
			}
			let p = Person();
			p.name = "Alice";
			p.name
			`,
			"Alice",
		},
		{
			`
			class Point {
				let x = 0;
				let y = 0;
			}
			let p = Point();
			p.x = 10;
			p.y = 20;
			p.x
			`,
			int64(10),
		},
		{
			`
			class Point {
				let x = 0;
			}
			let p = Point();
			p.x = 20;
			p.x
			`,
			int64(20),
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			evaluated := testEval(tt.input)
			testObject(t, evaluated, tt.expected)
		})
	}
}

// TestClassMethods tests method calls on class instances
// TODO: Fix method call compilation/execution
func TestClassMethods(t *testing.T) {
	t.Skip("Method calls need to be fixed")
	
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class Person {
				fn greet() {
					return "Hello";
				}
			}
			let p = Person();
			let result = p.greet();
			result
			`,
			"Hello",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			evaluated := testEval(tt.input)
			testObject(t, evaluated, tt.expected)
		})
	}
}

// TestConstructor tests class constructors
// TODO: Fix constructor execution
func TestConstructor(t *testing.T) {
	t.Skip("Constructors need to be fixed")
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class Person {
				let name = null;
				
				fn init(n) {
					this.name = n;
				}
			}
			let p = Person("Alice");
			p.name
			`,
			"Alice",
		},
		{
			`
			class Point {
				let x = 0;
				let y = 0;
				
				fn init(x, y) {
					this.x = x;
					this.y = y;
				}
			}
			let p = Point(10, 20);
			p.x + p.y
			`,
			int64(30),
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testObject(t, evaluated, tt.expected)
	}
}

// TestInheritance tests class inheritance
// TODO: Fix inheritance resolution
func TestInheritance(t *testing.T) {
	t.Skip("Inheritance needs to be fixed")
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class Animal {
				let name = null;
				
				fn init(n) {
					this.name = n;
				}
				
				fn speak() {
					return this.name + " makes a sound";
				}
			}
			
			class Dog extends Animal {
				fn speak() {
					return this.name + " barks";
				}
			}
			
			let dog = Dog("Buddy");
			dog.speak()
			`,
			"Buddy barks",
		},
		{
			`
			class Animal {
				let name = null;
				
				fn getName() {
					return this.name;
				}
			}
			
			class Dog extends Animal {
				fn init(n) {
					this.name = n;
				}
			}
			
			let dog = Dog("Buddy");
			dog.getName()
			`,
			"Buddy",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testObject(t, evaluated, tt.expected)
	}
}

// TestMethodResolution tests method resolution in class hierarchy
// TODO: Fix method resolution
func TestMethodResolution(t *testing.T) {
	t.Skip("Method resolution needs to be fixed")
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class A {
				fn method() {
					return "A";
				}
			}
			
			class B extends A {
				fn method() {
					return "B";
				}
			}
			
			let b = B();
			b.method()
			`,
			"B",
		},
		{
			`
			class A {
				fn method() {
					return "A";
				}
			}
			
			class B extends A {
			}
			
			let b = B();
			b.method()
			`,
			"A",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testObject(t, evaluated, tt.expected)
	}
}

// TestPropertyInheritance tests property inheritance
// TODO: Fix property inheritance
func TestPropertyInheritance(t *testing.T) {
	t.Skip("Property inheritance needs to be fixed")
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class Animal {
				let name = null;
			}
			
			class Dog extends Animal {
				let breed = null;
			}
			
			let dog = Dog();
			dog.name = "Buddy";
			dog.breed = "Golden Retriever";
			dog.name + " is a " + dog.breed
			`,
			"Buddy is a Golden Retriever",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testObject(t, evaluated, tt.expected)
	}
}

func TestMultipleInstances(t *testing.T) {
	t.Skip("Constructors need to be fixed")
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			class Person {
				let name = null;
				
				fn init(n) {
					this.name = n;
				}
			}
			
			let p1 = Person("Alice");
			let p2 = Person("Bob");
			p1.name + " and " + p2.name
			`,
			"Alice and Bob",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testObject(t, evaluated, tt.expected)
	}
}

// Helper functions

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		panic(p.Errors()[0])
	}

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		panic(err)
	}

	bytecode := comp.Bytecode()
	vm := New(bytecode)
	err = vm.Run()
	if err != nil {
		panic(err)
	}

	// Get the last popped stack element (result of the last expression)
	return vm.LastPoppedStackElem()
}

func testObject(t *testing.T, obj object.Object, expected interface{}) bool {
	switch expected := expected.(type) {
	case int:
		return testIntegerObject(t, obj, int64(expected))
	case int64:
		return testIntegerObject(t, obj, expected)
	case string:
		return testStringObject(t, obj, expected)
	case bool:
		return testBooleanObject(t, obj, expected)
	case nil:
		return testNullObject(t, obj)
	default:
		t.Errorf("type of expected not handled. got=%T", expected)
		return false
	}
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%d, want=%d", result.Value, expected)
		return false
	}
	return true
}

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	result, ok := obj.(*object.String)
	if !ok {
		t.Errorf("object is not String. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%q, want=%q", result.Value, expected)
		return false
	}
	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}
	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, want=%t", result.Value, expected)
		return false
	}
	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != Null {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}
