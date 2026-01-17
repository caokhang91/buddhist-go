package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/caokhang91/buddhist-go/pkg/compiler"
	"github.com/caokhang91/buddhist-go/pkg/lexer"
	"github.com/caokhang91/buddhist-go/pkg/object"
	"github.com/caokhang91/buddhist-go/pkg/parser"
	"github.com/caokhang91/buddhist-go/pkg/vm"
)

const VERSION = "1.0.0"

const BANNER = `
╔══════════════════════════════════════════════════════════════╗
║     ____            _     _ _     _     _                    ║
║    |  _ \          | |   | | |   (_)   | |                   ║
║    | |_) |_   _  __| | __| | |__  _ ___| |_                  ║
║    |  _ <| | | |/ _` + "`" + ` |/ _` + "`" + ` | '_ \| / __| __|                 ║
║    | |_) | |_| | (_| | (_| | | | | \__ \ |_                  ║
║    |____/ \__,_|\__,_|\__,_|_| |_|_|___/\__|                 ║
║                                                              ║
║               Go-Powered Interpreter Language                ║
║                     Version %s                            ║
╚══════════════════════════════════════════════════════════════╝
`

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		// Start REPL
		fmt.Printf(BANNER, VERSION)
		fmt.Println("\nType 'help' for commands, 'exit' to quit.")
		fmt.Println()
		startREPL(os.Stdin, os.Stdout)
	} else if args[0] == "-h" || args[0] == "--help" {
		printHelp()
	} else if args[0] == "-v" || args[0] == "--version" {
		fmt.Printf("Buddhist Lang version %s\n", VERSION)
	} else {
		// Execute file
		filename := args[0]
		if err := executeFile(filename); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	}
}

func printHelp() {
	fmt.Printf("Buddhist Lang - Go-Powered Interpreter Language (v%s)\n\n", VERSION)
	fmt.Println("Usage:")
	fmt.Println("  buddhist-go              Start the interactive REPL")
	fmt.Println("  buddhist-go <file>       Execute a script file")
	fmt.Println("  buddhist-go -h, --help   Show this help message")
	fmt.Println("  buddhist-go -v, --version Show version information")
	fmt.Println()
	fmt.Println("REPL Commands:")
	fmt.Println("  help     Show available commands")
	fmt.Println("  clear    Clear the screen")
	fmt.Println("  exit     Exit the REPL")
	fmt.Println()
	fmt.Println("Language Features:")
	fmt.Println("  - Variables: let x = 5; const PI = 3.14;")
	fmt.Println("  - Functions: fn add(a, b) { return a + b; }")
	fmt.Println("  - Conditionals: if (x > 5) { ... } else { ... }")
	fmt.Println("  - Loops: while (x < 10) { ... }")
	fmt.Println("  - Arrays: [1, 2, 3]")
	fmt.Println("  - Hash Maps: {\"key\": \"value\"}")
	fmt.Println("  - Concurrency: spawn fn() { ... }")
	fmt.Println("  - Channels: let ch = channel; ch <- value;")
}

func executeFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read file %s: %w", filename, err)
	}

	return execute(string(content))
}

func execute(input string) error {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "Parse error: %s\n", err)
		}
		return fmt.Errorf("parsing failed")
	}

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		return fmt.Errorf("compilation error: %w", err)
	}

	machine := vm.New(comp.Bytecode())
	err = machine.Run()
	if err != nil {
		return fmt.Errorf("runtime error: %w", err)
	}

	return nil
}

const PROMPT = ">>> "
const CONTINUE_PROMPT = "... "

func startREPL(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	constants := []object.Object{}
	globals := make([]object.Object, vm.GlobalsSize)
	symbolTable := compiler.NewSymbolTable()

	// Define builtins
	for i, v := range object.Builtins {
		symbolTable.DefineBuiltin(i, v.Name)
	}

	var inputBuffer strings.Builder

	for {
		if inputBuffer.Len() == 0 {
			fmt.Fprint(out, PROMPT)
		} else {
			fmt.Fprint(out, CONTINUE_PROMPT)
		}

		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()

		// Handle special commands
		if inputBuffer.Len() == 0 {
			switch strings.TrimSpace(line) {
			case "exit", "quit":
				fmt.Fprintln(out, "Goodbye!")
				return
			case "help":
				printREPLHelp(out)
				continue
			case "clear":
				fmt.Fprint(out, "\033[H\033[2J")
				continue
			}
		}

		inputBuffer.WriteString(line)
		inputBuffer.WriteString("\n")

		input := inputBuffer.String()

		// Try to parse
		l := lexer.New(input)
		p := parser.New(l)
		program := p.ParseProgram()

		// Check for incomplete input (might need more)
		if hasIncompleteInput(p.Errors()) {
			continue
		}

		// Reset input buffer
		inputBuffer.Reset()

		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		comp := compiler.NewWithState(symbolTable, constants)
		err := comp.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Compilation error: %s\n", err)
			continue
		}

		code := comp.Bytecode()
		constants = code.Constants

		machine := vm.NewWithGlobalsStore(code, globals)
		err = machine.Run()
		if err != nil {
			fmt.Fprintf(out, "Runtime error: %s\n", err)
			continue
		}

		lastPopped := machine.LastPoppedStackElem()
		if lastPopped != nil && lastPopped != vm.Null {
			io.WriteString(out, lastPopped.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

func hasIncompleteInput(errors []string) bool {
	for _, err := range errors {
		if strings.Contains(err, "expected next token") {
			return true
		}
	}
	return false
}

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "  ")
		io.WriteString(out, msg)
		io.WriteString(out, "\n")
	}
}

func printREPLHelp(out io.Writer) {
	fmt.Fprintln(out, "Available commands:")
	fmt.Fprintln(out, "  help     - Show this help message")
	fmt.Fprintln(out, "  clear    - Clear the screen")
	fmt.Fprintln(out, "  exit     - Exit the REPL")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Example expressions:")
	fmt.Fprintln(out, "  1 + 2 * 3")
	fmt.Fprintln(out, "  let x = 10")
	fmt.Fprintln(out, "  fn add(a, b) { a + b }")
	fmt.Fprintln(out, "  add(5, 3)")
	fmt.Fprintln(out, "  [1, 2, 3][0]")
	fmt.Fprintln(out, "  {\"name\": \"Buddhist\"}[\"name\"]")
	fmt.Fprintln(out, "  println(\"Hello, World!\")")
}
