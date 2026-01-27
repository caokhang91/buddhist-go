// AddressManager is a native app that runs the Address Management GUI script.
// By default it uses the embedded script; -script overrides with a file.
//
// Build (from repo root):
//   go build -ldflags="-s -w" -o AddressManager ./cmd/address-manager
// Windows:
//   GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o AddressManager.exe ./cmd/address-manager
package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	"github.com/caokhang91/buddhist-go/pkg/compiler"
	"github.com/caokhang91/buddhist-go/pkg/lexer"
	"github.com/caokhang91/buddhist-go/pkg/parser"
	"github.com/caokhang91/buddhist-go/pkg/vm"
)

//go:embed script.bl
var embeddedScript []byte

func main() {
	guiMode := flag.Bool("gui", true, "Run in GUI mode (default true for native app)")
	scriptPath := flag.String("script", "", "Path to script file (overrides embedded)")
	flag.Parse()

	var scriptContent string
	if *scriptPath != "" {
		data, err := os.ReadFile(*scriptPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading script %s: %v\n", *scriptPath, err)
			os.Exit(1)
		}
		scriptContent = string(data)
	} else {
		scriptContent = string(embeddedScript)
	}

	if !*guiMode {
		fmt.Fprintln(os.Stderr, "AddressManager is a GUI app; -gui=false is ignored, running with GUI.")
	}

	if err := runScript(scriptContent); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runScript(source string) error {
	l := lexer.NewOptimized(source)
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		for _, err := range p.Errors() {
			fmt.Fprintf(os.Stderr, "Parse error: %s\n", err)
		}
		return fmt.Errorf("parsing failed")
	}
	comp := compiler.New()
	if err := comp.Compile(program); err != nil {
		return fmt.Errorf("compilation error: %w", err)
	}
	bytecode := comp.Bytecode()
	machine := vm.New(bytecode)
	return machine.Run()
}
