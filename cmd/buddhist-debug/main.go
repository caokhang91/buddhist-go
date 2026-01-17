package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/caokhang91/buddhist-go/pkg/ast"
	"github.com/caokhang91/buddhist-go/pkg/compiler"
	"github.com/caokhang91/buddhist-go/pkg/lexer"
	"github.com/caokhang91/buddhist-go/pkg/parser"
	"github.com/caokhang91/buddhist-go/pkg/token"
	"github.com/caokhang91/buddhist-go/pkg/vm"
)

type DebugRequest struct {
	Command string                 `json:"command"`
	Args    map[string]interface{} `json:"args,omitempty"`
}

type DebugResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

type Breakpoint struct {
	ID       int
	File     string
	Line     int
	Verified bool
}

type DebugVM struct {
	vm            *vm.VM
	bytecode      *compiler.Bytecode
	breakpoints   map[string][]int // file -> lines
	stopped       bool
	stepMode      string // "none", "step", "stepIn", "stepOut"
	currentLine   int
	currentFile   string
	callDepth     int
	targetDepth   int
	mu            sync.Mutex
	conn          net.Conn
	lastFrame     *vm.Frame
	sourceLines   map[int]int // instruction offset -> line number
}

var debugPort = flag.Int("port", 2345, "Debug server port")
var debugFlag = flag.Bool("debug", false, "Enable debug mode")

func main() {
	flag.Parse()

	if !*debugFlag {
		fmt.Fprintf(os.Stderr, "This is a debug server. Use --debug flag.\n")
		os.Exit(1)
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: buddhist-debug --debug --port <port> <file>\n")
		os.Exit(1)
	}

	filename := args[0]
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Parse and compile
	l := lexer.NewOptimized(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "Parse error: %s\n", err)
		}
		os.Exit(1)
	}

	comp := compiler.New()
	err = comp.Compile(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation error: %v\n", err)
		os.Exit(1)
	}

	bytecode := comp.Bytecode()

	// Build source line mapping (simplified - would need proper source map in production)
	sourceLines := buildSourceLineMap(program, bytecode)

	// Start debug server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *debugPort))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error starting debug server: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Fprintf(os.Stderr, "Debug server listening on port %d\n", *debugPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection: %v\n", err)
			continue
		}

		debugVM := &DebugVM{
			vm:          vm.New(bytecode),
			bytecode:    bytecode,
			breakpoints: make(map[string][]int),
			stopped:     false,
			stepMode:    "none",
			currentFile: filename,
			sourceLines: sourceLines,
			conn:        conn,
		}

		go handleDebugConnection(debugVM)
	}
}

func buildSourceLineMap(program *ast.Program, bytecode *compiler.Bytecode) map[int]int {
	// Simplified source line mapping
	// In production, this would be built during compilation
	lines := make(map[int]int)
	
	// This is a placeholder - a real implementation would track
	// instruction offsets to source lines during compilation
	// For now, return empty map as this requires compiler modifications
	_ = program
	_ = bytecode
	
	return lines
}

func getTokenFromNode(node ast.Node) *token.Token {
	switch n := node.(type) {
	case *ast.ExpressionStatement:
		return &n.Token
	case *ast.LetStatement:
		return &n.Token
	case *ast.ReturnStatement:
		return &n.Token
	default:
		return nil
	}
}

func handleDebugConnection(d *DebugVM) {
	defer d.conn.Close()

	scanner := bufio.NewScanner(d.conn)
	writer := bufio.NewWriter(d.conn)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req DebugRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(writer, "Invalid request: "+err.Error())
			continue
		}

		var resp DebugResponse
		switch req.Command {
		case "setBreakpoints":
			resp = d.handleSetBreakpoints(req.Args)
		case "continue":
			resp = d.handleContinue()
		case "next":
			resp = d.handleNext()
		case "stepIn":
			resp = d.handleStepIn()
		case "stepOut":
			resp = d.handleStepOut()
		case "pause":
			resp = d.handlePause()
		case "stackTrace":
			resp = d.handleStackTrace()
		case "variables":
			resp = d.handleVariables(req.Args)
		case "evaluate":
			resp = d.handleEvaluate(req.Args)
		case "setVariable":
			resp = d.handleSetVariable(req.Args)
		case "disconnect":
			sendResponse(writer, DebugResponse{Success: true})
			return
		default:
			resp = DebugResponse{
				Success: false,
				Error:   "Unknown command: " + req.Command,
			}
		}

		sendResponse(writer, resp)
		writer.Flush()
	}
}

func (d *DebugVM) handleSetBreakpoints(args map[string]interface{}) DebugResponse {
	d.mu.Lock()
	defer d.mu.Unlock()

	file, _ := args["file"].(string)
	breakpoints, _ := args["breakpoints"].([]interface{})

	lines := make([]int, 0)
	for _, bp := range breakpoints {
		if bpMap, ok := bp.(map[string]interface{}); ok {
			if line, ok := bpMap["line"].(float64); ok {
				lines = append(lines, int(line))
			}
		}
	}

	d.breakpoints[file] = lines

	return DebugResponse{
		Success: true,
		Data: map[string]interface{}{
			"breakpoints": breakpoints,
		},
	}
}

func (d *DebugVM) handleContinue() DebugResponse {
	d.mu.Lock()
	d.stopped = false
	d.stepMode = "none"
	d.mu.Unlock()

	// Run VM until breakpoint or end
	go d.runVM()

	return DebugResponse{Success: true}
}

func (d *DebugVM) handleNext() DebugResponse {
	d.mu.Lock()
	d.stopped = false
	d.stepMode = "step"
	d.mu.Unlock()

	go d.runVM()

	return DebugResponse{Success: true}
}

func (d *DebugVM) handleStepIn() DebugResponse {
	d.mu.Lock()
	d.stopped = false
	d.stepMode = "stepIn"
	d.targetDepth = d.callDepth
	d.mu.Unlock()

	go d.runVM()

	return DebugResponse{Success: true}
}

func (d *DebugVM) handleStepOut() DebugResponse {
	d.mu.Lock()
	d.stopped = false
	d.stepMode = "stepOut"
	d.targetDepth = d.callDepth - 1
	d.mu.Unlock()

	go d.runVM()

	return DebugResponse{Success: true}
}

func (d *DebugVM) handlePause() DebugResponse {
	d.mu.Lock()
	d.stopped = true
	d.mu.Unlock()

	return DebugResponse{Success: true}
}

func (d *DebugVM) handleStackTrace() DebugResponse {
	d.mu.Lock()
	defer d.mu.Unlock()

	frames := []map[string]interface{}{
		{
			"id":     0,
			"name":   "main",
			"line":   d.currentLine,
			"column": 1,
			"source": d.currentFile,
		},
	}

	return DebugResponse{
		Success: true,
		Data: map[string]interface{}{
			"frames": frames,
		},
	}
}

func (d *DebugVM) handleVariables(args map[string]interface{}) DebugResponse {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Get variables from VM stack and globals
	variables := []map[string]interface{}{}

	// This is simplified - would need to inspect VM state
	// For now, return empty variables list

	return DebugResponse{
		Success: true,
		Data: map[string]interface{}{
			"variables": variables,
		},
	}
}

func (d *DebugVM) handleEvaluate(args map[string]interface{}) DebugResponse {
	// Simplified evaluation - would need to parse and evaluate expression
	expr, _ := args["expression"].(string)

	return DebugResponse{
		Success: true,
		Data: map[string]interface{}{
			"result": "Evaluation not yet implemented: " + expr,
		},
	}
}

func (d *DebugVM) handleSetVariable(args map[string]interface{}) DebugResponse {
	// Simplified - would need to update VM state
	return DebugResponse{
		Success: false,
		Error:   "Set variable not yet implemented",
	}
}

func (d *DebugVM) runVM() {
	// This would need to be integrated with VM.Run() to support stepping
	// For now, just run the VM
	err := d.vm.Run()
	if err != nil {
		sendEvent(d.conn, "stopped", map[string]interface{}{
			"reason": "error",
			"error":  err.Error(),
		})
	} else {
		sendEvent(d.conn, "stopped", map[string]interface{}{
			"reason": "end",
		})
	}
}

func sendResponse(writer *bufio.Writer, resp DebugResponse) {
	data, _ := json.Marshal(resp)
	writer.Write(data)
	writer.WriteString("\n")
}

func sendError(writer *bufio.Writer, errMsg string) {
	sendResponse(writer, DebugResponse{
		Success: false,
		Error:   errMsg,
	})
}

func sendEvent(conn net.Conn, event string, data map[string]interface{}) {
	eventData := map[string]interface{}{
		"event": event,
	}
	for k, v := range data {
		eventData[k] = v
	}

	resp := DebugResponse{
		Success: true,
		Data:    eventData,
	}

	writer := bufio.NewWriter(conn)
	sendResponse(writer, resp)
	writer.Flush()
}
