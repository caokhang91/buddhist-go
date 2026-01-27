package main

import (
	"bytes"
	"bufio"
	"fmt"
	"html"
	"io"
	"os"
	"strings"
	"time"

	"github.com/caokhang91/buddhist-go/pkg/compiler"
	"github.com/caokhang91/buddhist-go/pkg/lexer"
	"github.com/caokhang91/buddhist-go/pkg/object"
	"github.com/caokhang91/buddhist-go/pkg/parser"
	"github.com/caokhang91/buddhist-go/pkg/vm"
)

const VERSION = "1.0.3"

func isGUIFlag(s string) bool { return s == "--gui" || s == "-g" }

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
	} else if args[0] == "--benchmark" || args[0] == "-b" {
		// Benchmark mode
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: buddhist-go --benchmark <file>")
			os.Exit(1)
		}
		if err := benchmarkFile(args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	} else if isGUIFlag(args[0]) {
		// GUI mode: --gui <file> or -g <file>
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: buddhist-go --gui <file>")
			fmt.Fprintln(os.Stderr, "  Run a script with GUI; stdout/stderr go to terminal, no HTML output.")
			os.Exit(1)
		}
		filename := args[1]
		if isGUIFlag(filename) {
			fmt.Fprintln(os.Stderr, "Usage: buddhist-go --gui <file>")
			fmt.Fprintln(os.Stderr, "  Provide a .bl script path, not a flag.")
			os.Exit(1)
		}
		if err := runFileDirect(filename); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	} else if len(args) >= 2 && isGUIFlag(args[1]) {
		// <file> --gui or <file> -g
		filename := args[0]
		if err := runFileDirect(filename); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			os.Exit(1)
		}
	} else {
		// Execute file (output wrapped as HTML)
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
	fmt.Println("  buddhist-go                    Start the interactive REPL")
	fmt.Println("  buddhist-go <file>             Execute a script file")
	fmt.Println("  buddhist-go -h, --help         Show this help message")
	fmt.Println("  buddhist-go -v, --version      Show version information")
	fmt.Println("  buddhist-go -b, --benchmark <file>  Benchmark a script file")
	fmt.Println("  buddhist-go -g, --gui <file>       Run script with GUI (direct stdout/stderr, no HTML)")
	fmt.Println()
	fmt.Println("Performance Features (v1.0.0):")
	fmt.Println("  - Optimized VM with inlined operations and cached frame references")
	fmt.Println("  - Fast paths for integer arithmetic and comparisons")
	fmt.Println("  - Small integer caching (-128 to 256)")
	fmt.Println("  - Constant folding at compile time")
	fmt.Println("  - String interning for memory efficiency")
	fmt.Println("  - Optimized lexer with byte slice processing")
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

// runFileDirect runs a script file without capturing stdout/stderr and without HTML.
// Use for GUI scripts: ./buddhist-go --gui examples/address_management.bl
func runFileDirect(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read file %s: %w", filename, err)
	}
	return execute(string(content))
}

func executeFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read file %s: %w", filename, err)
	}

	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	
	stdoutR, stdoutW, _ := os.Pipe()
	stderrR, stderrW, _ := os.Pipe()
	
	os.Stdout = stdoutW
	os.Stderr = stderrW
	
	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutDone := make(chan struct{})
	stderrDone := make(chan struct{})
	
	// Capture stdout in a goroutine
	go func() {
		io.Copy(&stdoutBuf, stdoutR)
		close(stdoutDone)
	}()
	
	// Capture stderr in a goroutine
	go func() {
		io.Copy(&stderrBuf, stderrR)
		close(stderrDone)
	}()
	
	// Execute the script
	execErr := execute(string(content))
	
	// Close the write pipes to signal end of output
	stdoutW.Close()
	stderrW.Close()
	
	// Restore original stdout/stderr
	os.Stdout = oldStdout
	os.Stderr = oldStderr
	
	// Wait for capture to complete
	<-stdoutDone
	<-stderrDone
	
	// Format and output as HTML
	htmlOutput := formatAsHTML(stdoutBuf.Bytes(), stderrBuf.Bytes(), execErr)
	fmt.Print(htmlOutput)
	
	return execErr
}

func execute(input string) error {
	// Use optimized lexer
	l := lexer.NewOptimized(input)
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

	bytecode := comp.Bytecode()
	machine := vm.New(bytecode)
	err = machine.Run()
	if err != nil {
		return fmt.Errorf("runtime error: %w", err)
	}

	return nil
}

// benchmarkFile runs a file multiple times and reports timing
func benchmarkFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("could not read file %s: %w", filename, err)
	}

	input := string(content)
	iterations := 10

	fmt.Printf("Benchmarking %s (%d iterations)\n", filename, iterations)
	fmt.Println("----------------------------------------")

	// Warm up
	for i := 0; i < 3; i++ {
		executeBenchmark(input)
	}

	// Run benchmark
	var total time.Duration
	var minTime, maxTime time.Duration
	for i := 0; i < iterations; i++ {
		start := time.Now()
		err := executeBenchmark(input)
		elapsed := time.Since(start)
		if err != nil {
			return err
		}
		total += elapsed
		if i == 0 || elapsed < minTime {
			minTime = elapsed
		}
		if elapsed > maxTime {
			maxTime = elapsed
		}
	}
	avg := total / time.Duration(iterations)

	fmt.Printf("Average: %v\n", avg)
	fmt.Printf("Min:     %v\n", minTime)
	fmt.Printf("Max:     %v\n", maxTime)
	fmt.Printf("Total:   %v\n", total)

	return nil
}

// executeBenchmark runs the code without printing output
func executeBenchmark(input string) error {
	l := lexer.NewOptimized(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		return fmt.Errorf("parsing failed")
	}

	comp := compiler.New()
	err := comp.Compile(program)
	if err != nil {
		return err
	}

	bytecode := comp.Bytecode()
	machine := vm.New(bytecode)
	return machine.Run()
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

		// Check for unbalanced braces first - if braces aren't balanced,
		// we definitely need more input
		if countBraces(input) > 0 {
			continue
		}

		// Try to parse using optimized lexer
		l := lexer.NewOptimized(input)
		p := parser.New(l)
		program := p.ParseProgram()

		// Check for incomplete input (might need more)
		if hasIncompleteInput(p.Errors(), input) {
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

func hasIncompleteInput(errors []string, input string) bool {
	for _, err := range errors {
		if strings.Contains(err, "expected next token") {
			return true
		}
		// If we see an error about unexpected EOF or missing closing brace
		// and we have unbalanced input, we need more
		if strings.Contains(err, "unexpected EOF") ||
			strings.Contains(err, "expected }") ||
			strings.Contains(err, "expected )") ||
			strings.Contains(err, "expected ]") {
			return true
		}
	}
	// Also check if the input ends with an operator or opening brace
	// which typically indicates incomplete input
	trimmed := strings.TrimRight(input, " \t\n")
	if len(trimmed) > 0 {
		lastChar := trimmed[len(trimmed)-1]
		if lastChar == '{' || lastChar == '(' || lastChar == '[' ||
			lastChar == ',' || lastChar == '+' || lastChar == '-' ||
			lastChar == '*' || lastChar == '/' || lastChar == '=' ||
			lastChar == '<' || lastChar == '>' || lastChar == '&' ||
			lastChar == '|' || lastChar == ':' {
			return true
		}
	}
	return false
}

// countBraces returns the net count of open braces (open - close)
// Also accounts for braces inside strings and comments
func countBraces(input string) int {
	count := 0
	inString := false
	inComment := false
	stringChar := byte(0)
	
	for i := 0; i < len(input); i++ {
		ch := input[i]
		
		// Handle line comments
		if !inString && i+1 < len(input) && ch == '/' && input[i+1] == '/' {
			// Skip to end of line
			for i < len(input) && input[i] != '\n' {
				i++
			}
			continue
		}
		
		// Handle block comments
		if !inString && i+1 < len(input) && ch == '/' && input[i+1] == '*' {
			inComment = true
			i++
			continue
		}
		if inComment && i+1 < len(input) && ch == '*' && input[i+1] == '/' {
			inComment = false
			i++
			continue
		}
		if inComment {
			continue
		}
		
		// Handle strings
		if !inString && (ch == '"' || ch == '\'') {
			inString = true
			stringChar = ch
			continue
		}
		if inString && ch == stringChar {
			// Check for escape
			escapes := 0
			for j := i - 1; j >= 0 && input[j] == '\\'; j-- {
				escapes++
			}
			if escapes%2 == 0 {
				inString = false
			}
			continue
		}
		
		// Count braces
		if !inString {
			if ch == '{' || ch == '(' || ch == '[' {
				count++
			} else if ch == '}' || ch == ')' || ch == ']' {
				count--
			}
		}
	}
	
	return count
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

// formatAsHTML wraps the captured output in a styled HTML document
func formatAsHTML(stdout []byte, stderr []byte, execErr error) string {
	var buf strings.Builder
	
	buf.WriteString("<!DOCTYPE html>\n")
	buf.WriteString("<html lang=\"en\">\n")
	buf.WriteString("<head>\n")
	buf.WriteString("  <meta charset=\"UTF-8\">\n")
	buf.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	buf.WriteString("  <title>Buddhist Script Output</title>\n")
	buf.WriteString("  <style>\n")
	buf.WriteString("    * { margin: 0; padding: 0; box-sizing: border-box; }\n")
	buf.WriteString("    body { font-family: 'Consolas', 'Monaco', 'Courier New', monospace; background: #1e1e1e; color: #d4d4d4; line-height: 1.6; padding: 20px; }\n")
	buf.WriteString("    .container { max-width: 1200px; margin: 0 auto; background: #252526; border-radius: 8px; padding: 20px; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.3); }\n")
	buf.WriteString("    h1 { color: #4ec9b0; margin-bottom: 20px; font-size: 24px; border-bottom: 2px solid #3c3c3c; padding-bottom: 10px; }\n")
	buf.WriteString("    .output-section { margin-bottom: 20px; }\n")
	buf.WriteString("    .output-section h2 { color: #569cd6; font-size: 18px; margin-bottom: 10px; }\n")
	buf.WriteString("    pre { background: #1e1e1e; padding: 15px; border-radius: 4px; overflow-x: auto; border: 1px solid #3c3c3c; white-space: pre-wrap; word-wrap: break-word; }\n")
	buf.WriteString("    .stdout { color: #d4d4d4; }\n")
	buf.WriteString("    .stderr { color: #f48771; background: #2d1b1b; border-color: #5a1f1f; }\n")
	buf.WriteString("    .error { color: #f48771; background: #2d1b1b; padding: 15px; border-radius: 4px; border: 1px solid #5a1f1f; margin-top: 10px; }\n")
	buf.WriteString("    .empty { color: #808080; font-style: italic; }\n")
	buf.WriteString("    .meta { color: #808080; font-size: 12px; margin-bottom: 20px; }\n")
	buf.WriteString("  </style>\n")
	buf.WriteString("</head>\n")
	buf.WriteString("<body>\n")
	buf.WriteString("  <div class=\"container\">\n")
	buf.WriteString("    <h1>Buddhist Script Output</h1>\n")
	buf.WriteString("    <div class=\"meta\">Generated at " + time.Now().Format("2006-01-02 15:04:05") + "</div>\n")
	
	// Standard output
	buf.WriteString("    <div class=\"output-section\">\n")
	buf.WriteString("      <h2>Standard Output</h2>\n")
	if len(stdout) > 0 {
		buf.WriteString("      <pre class=\"stdout\">")
		buf.WriteString(html.EscapeString(string(stdout)))
		buf.WriteString("</pre>\n")
	} else {
		buf.WriteString("      <pre class=\"stdout empty\">(no output)</pre>\n")
	}
	buf.WriteString("    </div>\n")
	
	// Standard error (if any)
	if len(stderr) > 0 {
		buf.WriteString("    <div class=\"output-section\">\n")
		buf.WriteString("      <h2>Standard Error</h2>\n")
		buf.WriteString("      <pre class=\"stderr\">")
		buf.WriteString(html.EscapeString(string(stderr)))
		buf.WriteString("</pre>\n")
		buf.WriteString("    </div>\n")
	}
	
	// Execution error (if any)
	if execErr != nil {
		buf.WriteString("    <div class=\"output-section\">\n")
		buf.WriteString("      <h2>Execution Error</h2>\n")
		buf.WriteString("      <div class=\"error\">")
		buf.WriteString(html.EscapeString(execErr.Error()))
		buf.WriteString("</div>\n")
		buf.WriteString("    </div>\n")
	}
	
	buf.WriteString("  </div>\n")
	buf.WriteString("</body>\n")
	buf.WriteString("</html>\n")
	
	return buf.String()
}
