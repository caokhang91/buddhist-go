# Buddhist Language VSCode Extension

VSCode extension providing language support and debugging capabilities for the Buddhist programming language.

## Features

- ✅ Syntax highlighting for `.bl` files
- ✅ Code completion and IntelliSense
- ✅ Debugger support with breakpoints
- ✅ Step through, step into, step out
- ✅ Variable inspection
- ✅ Expression evaluation
- ✅ Call stack navigation

## Installation

### From Source

1. Clone the repository:
```bash
git clone https://github.com/caokhang91/buddhist-go.git
cd buddhist-go/vscode-extension
```

2. Install dependencies:
```bash
npm install
```

3. Compile the extension:
```bash
npm run compile
```

4. Press `F5` in VSCode to open a new window with the extension loaded, or package it:
```bash
npm install -g vsce
vsce package
```

5. Install the `.vsix` file in VSCode:
   - Open VSCode
   - Go to Extensions view (`Ctrl+Shift+X` or `Cmd+Shift+X`)
   - Click the `...` menu → "Install from VSIX..."
   - Select the generated `.vsix` file

## Building the Debug Server

The debug adapter requires a Go debug server. Build it with:

```bash
cd ../cmd/buddhist-debug
go build -o buddhist-debug
```

Make sure the `buddhist-debug` executable is in your PATH or update the extension configuration.

## Usage

### Debugging

1. Open a `.bl` file in VSCode
2. Set breakpoints by clicking in the gutter next to line numbers
3. Press `F5` or go to Run and Debug view
4. Select "Launch Buddhist Program" configuration
5. The debugger will start and stop at breakpoints

### Debug Configuration

Create a `.vscode/launch.json` file in your workspace:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "buddhist",
      "request": "launch",
      "name": "Launch Buddhist Program",
      "program": "${workspaceFolder}/examples/hello.bl",
      "stopOnEntry": false,
      "showDebugOutput": true
    }
  ]
}
```

### Configuration Options

- `program`: Path to the Buddhist program file to debug
- `args`: Command line arguments (array of strings)
- `cwd`: Working directory
- `env`: Environment variables (object)
- `stopOnEntry`: Stop at entry point (default: false)
- `showDebugOutput`: Show debug adapter output (default: false)

## Debug Features

### Breakpoints
- Click in the gutter to set/remove breakpoints
- Conditional breakpoints: Right-click on a breakpoint to add conditions
- Logpoints: Log messages without stopping execution

### Stepping
- **Continue** (`F5`): Continue execution until next breakpoint
- **Step Over** (`F10`): Execute current line and stop at next line
- **Step Into** (`F11`): Step into function calls
- **Step Out** (`Shift+F11`): Step out of current function

### Variable Inspection
- View variables in the Variables panel
- Hover over variables to see their values
- Use the Debug Console to evaluate expressions

### Call Stack
- Navigate the call stack in the Call Stack panel
- Click on frames to jump to that location

## Language Features

### Syntax Highlighting
- Keywords: `fn`, `let`, `const`, `if`, `else`, `return`, etc.
- Strings: Single and double quoted strings
- Numbers: Integers and floats
- Comments: Line (`//`) and block (`/* */`) comments

### Code Navigation
- Go to definition (when implemented)
- Find references (when implemented)
- Symbol search (when implemented)

## Development

### Project Structure

```
vscode-extension/
├── src/
│   ├── extension.ts          # Extension entry point
│   ├── debugAdapter.ts       # Debug adapter factory
│   └── debugSession.ts       # Debug session implementation
├── syntaxes/
│   └── buddhist.tmLanguage.json  # TextMate grammar
├── language-configuration.json   # Language configuration
├── package.json              # Extension manifest
└── tsconfig.json            # TypeScript configuration
```

### Building

```bash
npm run compile      # Compile TypeScript
npm run watch        # Watch mode for development
```

### Testing

```bash
npm test
```

## Requirements

- VSCode 1.80.0 or higher
- Node.js 18+ (for extension development)
- Go 1.24+ (for building debug server)
- Buddhist interpreter (`buddhist` command in PATH)

## Troubleshooting

### Debugger not starting
- Ensure `buddhist-debug` is built and in PATH
- Check that the program file path is correct
- Enable `showDebugOutput` to see debug adapter logs

### Breakpoints not working
- Ensure the file path matches exactly
- Check that breakpoints are set on executable lines (not comments/whitespace)

### Extension not loading
- Check VSCode Developer Console for errors (`Help` → `Toggle Developer Tools`)
- Ensure all dependencies are installed (`npm install`)
- Rebuild the extension (`npm run compile`)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

Same as the main Buddhist Language project (MIT License).
