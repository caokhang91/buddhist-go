# Quick Start Guide

## Prerequisites

1. **Node.js** (v18 or higher)
2. **Go** (v1.24 or higher)
3. **VSCode** (v1.80.0 or higher)
4. **Buddhist interpreter** (`buddhist` command in PATH)

## Setup Steps

### 1. Build the Debug Server

```bash
cd cmd/buddhist-debug
go build -o buddhist-debug
```

Make sure `buddhist-debug` is in your PATH or update the extension configuration.

### 2. Install Extension Dependencies

```bash
cd vscode-extension
npm install
```

### 3. Compile the Extension

```bash
npm run compile
```

### 4. Install the Extension

#### Option A: Development Mode
- Open the `vscode-extension` folder in VSCode
- Press `F5` to launch a new VSCode window with the extension loaded

#### Option B: Package and Install
```bash
npm install -g vsce
vsce package
```

Then install the generated `.vsix` file:
- Open VSCode
- Go to Extensions (`Ctrl+Shift+X` or `Cmd+Shift+X`)
- Click `...` → "Install from VSIX..."
- Select the `.vsix` file

## Using the Debugger

1. Open a `.bl` file
2. Set breakpoints by clicking in the gutter
3. Press `F5` or go to Run and Debug view
4. Select "Launch Buddhist Program"
5. The debugger will start

## Configuration

Create `.vscode/launch.json` in your workspace:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "type": "buddhist",
      "request": "launch",
      "name": "Launch Buddhist Program",
      "program": "${workspaceFolder}/examples/hello.bl"
    }
  ]
}
```

## Troubleshooting

- **Debugger won't start**: Check that `buddhist-debug` is built and accessible
- **Breakpoints not working**: Ensure file paths match exactly
- **Extension errors**: Check VSCode Developer Console (`Help` → `Toggle Developer Tools`)
