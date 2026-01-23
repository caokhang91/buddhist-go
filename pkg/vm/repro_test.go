package vm

import (
	"testing"
	"github.com/caokhang91/buddhist-go/pkg/lexer"
	"github.com/caokhang91/buddhist-go/pkg/parser"
	"github.com/caokhang91/buddhist-go/pkg/compiler"
	"github.com/caokhang91/buddhist-go/pkg/object"
)

func TestFibonacci(t *testing.T) {
	input := `
	place fib = fn(n) {
		if (n <= 1) {
			return n;
		}
		return fib(n - 1) + fib(n - 2);
	};
	place i = 0;
	while (i < 3) {
		fib(i);
		i = i + 1;
	}
	i;
	`
	
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	
	c := compiler.New()
	err := c.Compile(program)
	if err != nil {
		t.Fatalf("compiler error: %s", err)
	}
	
	vm := New(c.Bytecode())
	err = vm.Run()
	if err != nil {
		t.Fatalf("vm error: %s", err)
	}
	
	result := vm.LastPoppedStackElem()
	testIntegerObject(t, result, 3)
}


