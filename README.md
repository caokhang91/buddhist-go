# Buddhist Lang - Go-Powered Interpreter Language

[![Build](https://github.com/caokhang91/buddhist-go/actions/workflows/build.yml/badge.svg)](https://github.com/caokhang91/buddhist-go/actions/workflows/build.yml)
[![Release](https://github.com/caokhang91/buddhist-go/actions/workflows/release.yml/badge.svg)](https://github.com/caokhang91/buddhist-go/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A high-performance bytecode interpreter language built with Go, leveraging Go's runtime for concurrency via Goroutines and Channels.

## Features

- **High Performance**: Optimized bytecode VM with cached frame references
- **Concurrency Support**: Native `spawn` keyword and channels for concurrent programming
- **PHP-Style Arrays**: Ordered hash maps with O(1) lookup and maintained insertion order
- **Constant Folding**: Compile-time optimization for constant expressions
- **Integer Caching**: Pre-allocated small integers (-128 to 256) to reduce GC pressure
- **Optimized Lexer**: Byte slice processing for faster tokenization
- **GUI Support**: Built-in GUI functions powered by [faiface/pixel](https://github.com/faiface/pixel) for creating windows and interactive applications

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/caokhang91/buddhist-go.git
cd buddhist-go

# Build the interpreter (default: no GUI, for CI/headless)
go build -o buddhist-go .

# Build with GUI support (requires CGO and OpenGL; for -g/--gui and AddressManager)
go build -tags gui -o buddhist-go .

# Or install directly (add -tags gui if you need GUI)
go install .
```

### From Release

Download pre-built binaries from the [Releases](https://github.com/caokhang91/buddhist-go/releases) page.

## Quick Start

### Interactive REPL

```bash
./buddhist-go
```

```
╔══════════════════════════════════════════════════════════════╗
║               Go-Powered Interpreter Language                ║
║                     Version 1.0.0                            ║
╚══════════════════════════════════════════════════════════════╝

>>> println("Hello, World!")
Hello, World!
>>> place x = 10
>>> x * 2
20
```

### Run a Script

```bash
./buddhist-go examples/hello.bl
```

### VS Code / Cursor Extension

Có extension cho VS Code và Cursor: syntax highlighting cho `.bl`, debug (breakpoints, step, biến). Cách thêm:

1. **Build interpreter và debug server** (từ repo root):
   ```bash
   go build -o buddhist-go .
   cd cmd/buddhist-debug && go build -o buddhist-debug
   ```
   Đảm bảo `buddhist-go` và `buddhist-debug` trong PATH (hoặc cấu hình `buddhist.interpreterPath` trong extension).

2. **Cài extension từ source**:
   ```bash
   cd vscode-extension
   npm install
   npm run compile
   ```
   - **Chạy thử**: mở thư mục `vscode-extension` trong VS Code/Cursor, nhấn **F5** → mở cửa sổ mới có extension.
   - **Cài bằng VSIX**: `npm install -g vsce && vsce package` → trong VS Code/Cursor: Extensions (`Ctrl+Shift+X` / `Cmd+Shift+X`) → `...` → **Install from VSIX...** → chọn file `.vsix` vừa tạo.

3. **Dùng debug**: mở file `.bl`, đặt breakpoint (click gutter), F5 → chọn **Launch Buddhist Program**. Cần `.vscode/launch.json` với `"type": "buddhist"`, `"program": "${workspaceFolder}/examples/hello.bl"` (xem [vscode-extension/README.md](vscode-extension/README.md) và [vscode-extension/QUICKSTART.md](vscode-extension/QUICKSTART.md)).

## Language Syntax

### Variables

```javascript
place x = 5;
const PI = 3.14;
set x = x + 1;
```

### Functions

```javascript
fn add(a, b) {
    return a + b;
}

// Anonymous functions
place multiply = fn(a, b) { a * b };

// Closures
fn counter() {
    place count = 0;
    return fn() {
        count = count + 1;
        return count;
    };
}
```

### Control Flow

```javascript
// If-else
if (x > 5) {
    println("greater");
} else {
    println("smaller or equal");
}

// While loop
while (x < 10) {
    println(x);
    x = x + 1;
}

// For loop
for (place i = 0; i < 10; i = i + 1) {
    println(i);
}
```

### Error handling

Use `try` / `catch` / `finally` and `throw` to handle errors. Built-in functions that fail (e.g. wrong types or invalid values) return error objects that are treated as throws, so you can catch them with `try`/`catch`.

```javascript
// Basic try/catch
try {
    place n = int("not a number");  // builtin returns error → thrown
} catch (e) {
    println("Caught: ", e);         // e is the error message/object
}

// try/catch with optional variable
try {
    throw "something went wrong";
} catch (err) {
    println(err);                   // "something went wrong"
}

// try/catch/finally
try {
    place x = first(null);          // type error from builtin
} catch (e) {
    println("Error: ", e);
} finally {
    println("cleanup runs either way");
}

// Uncaught throws become runtime errors and stop the script
throw "fatal";  // if not inside try, script exits with "uncaught throw"
```

| Construct   | Description |
|------------|-------------|
| `try { ... }` | Run code that may throw |
| `catch (e) { ... }` | Run if something was thrown; `e` is optional and receives the thrown value |
| `catch { ... }` | Catch without binding the thrown value |
| `finally { ... }` | Run after try/catch (on normal exit or after catch) |
| `throw expr;` | Throw a value (string, number, or any object); execution jumps to the nearest catch/finally |

### Arrays

```javascript
// Standard arrays
place arr = [1, 2, 3, 4, 5];
println(arr[0]);  // 1
println(len(arr)); // 5

// PHP-style arrays with keys
place map = [
    "name" => "Buddhist",
    "version" => "1.0.0",
    0 => "indexed"
];
println(map["name"]);  // Buddhist
```

### Hash Maps

```javascript
place person = {
    "name": "John",
    "age": 30
};
println(person["name"]);  // John
```

### Concurrency

```javascript
// Create a channel
place ch = channel;

// Spawn a goroutine
spawn fn() {
    ch <- "Hello from goroutine!";
};

// Receive from channel
place msg = <-ch;
println(msg);
```

## Built-in Functions

### I/O Functions

| Function | Description |
|----------|-------------|
| `println(...)` | Print values with newline |
| `print(...)` | Print values without newline |

### Type Functions

| Function | Description |
|----------|-------------|
| `len(x)` | Get length of array/string |
| `type(x)` | Get type of value |
| `str(x)` | Convert to string |
| `int(x)` | Convert to integer |
| `float(x)` | Convert to float |

### Array Functions

| Function | Description |
|----------|-------------|
| `first(arr)` | Get first element of array |
| `last(arr)` | Get last element of array |
| `rest(arr)` | Get array without first element |
| `push(arr, val)` | Append value to array |
| `slice(arr, start, end)` | Get a slice of array |
| `range(end)` or `range(start, end, step)` | Generate array of numbers |
| `map(arr, fn)` | Apply function to each element |
| `filter(arr, fn)` | Filter elements by predicate |
| `reduce(arr, fn, initial)` | Reduce array to single value |
| `reverse(arr)` | Reverse array |
| `concat(arr1, arr2, ...)` | Concatenate arrays |
| `contains(arr, val)` | Check if array contains value |
| `indexOf(arr, val)` | Get index of value in array |
| `unique(arr)` | Remove duplicate elements |
| `flatten(arr)` | Flatten nested arrays |
| `sum(arr)` | Sum of numeric array |
| `min(arr)` | Minimum value in array |
| `max(arr)` | Maximum value in array |
| `avg(arr)` | Average of numeric array |

### Math Functions

| Function | Description |
|----------|-------------|
| `sqrt(x)` | Square root |
| `pow(base, exp)` | Power function |
| `abs(x)` | Absolute value |
| `floor(x)` | Round down to integer |
| `ceil(x)` | Round up to integer |
| `round(x)` | Round to nearest integer |
| `sin(x)` | Sine (radians) |
| `cos(x)` | Cosine (radians) |
| `tan(x)` | Tangent (radians) |
| `log(x)` | Natural logarithm |
| `log10(x)` | Base-10 logarithm |
| `exp(x)` | Exponential (e^x) |
| `random()` | Random float 0-1 |
| `random(n)` | Random integer 0 to n-1 |
| `random(min, max)` | Random integer min to max |

### String Functions

| Function | Description |
|----------|-------------|
| `split(str, sep)` | Split string by separator |
| `join(arr, sep)` | Join array elements with separator |
| `trim(str)` | Remove leading/trailing whitespace |
| `trim(str, chars)` | Remove specific characters |
| `trimLeft(str)` | Remove leading whitespace |
| `trimRight(str)` | Remove trailing whitespace |
| `upper(str)` | Convert to uppercase |
| `lower(str)` | Convert to lowercase |
| `substring(str, start, end)` | Extract substring |
| `replace(str, old, new)` | Replace all occurrences |
| `replace(str, old, new, n)` | Replace first n occurrences |
| `startsWith(str, prefix)` | Check if string starts with prefix |
| `endsWith(str, suffix)` | Check if string ends with suffix |
| `repeat(str, n)` | Repeat string n times |

### GUI Functions

Built-in GUI functions powered by [faiface/pixel](https://github.com/faiface/pixel) for creating graphical applications. Chi tiết hiện trạng và kế hoạch phát triển: **[ROADMAP_GUI.md](ROADMAP_GUI.md)**.

| Function | Description |
|----------|-------------|
| `gui_window(config)` | Create a new GUI window with configuration (title, width, height, vsync) |
| `gui_button(window, config)` | Create a button with text, position, size, and onClick callback |
| `gui_table(window, config)` | Create a table with headers and data rows for displaying structured data |
| `gui_show(window)` | Mark a window to be shown (windows are created when `gui_run()` is called) |
| `gui_alert(window, message)` | Show a modal alert (message + OK) on the window |
| `gui_close(window)` | Close and remove a window |
| `gui_run()` | Start the GUI event loop (creates all windows and handles events) |

**Example:**

```javascript
// Create a window
place window = gui_window({
    "title": "My Buddhist App",
    "width": 800,
    "height": 600,
    "vsync": true
});

// Add a button with click handler
place btn = gui_button(window, {
    "text": "Click Me!",
    "x": 100.0,
    "y": 100.0,
    "width": 200.0,
    "height": 50.0,
    "onClick": fn() {
        println("Button was clicked!");
    }
});

// Show window and run event loop
gui_show(window);
gui_run();
```

**Table Example:**

```javascript
// Create a table for displaying data
place table = gui_table(window, {
    "x": 50.0,
    "y": 100.0,
    "width": 700.0,
    "height": 300.0,
    "headers": ["Name", "Address", "City"],
    "data": [
        ["John Doe", "123 Main St", "New York"],
        ["Jane Smith", "456 Oak Ave", "Los Angeles"]
    ],
    "rowHeight": 25.0,
    "headerHeight": 30.0,
    "onRowClick": fn(rowIndex) {
        println("Row clicked: ", rowIndex);
    }
});
```

**Running GUI scripts:** Use `--gui` or `-g` so the window opens and stdout goes to the terminal (without HTML output):
```bash
./buddhist-go --gui examples/address_management.bl
# or
./buddhist-go -g examples/gui_example.bl
```

**macOS: "Cocoa: Failed to find service port for display"** — GUI cần có display thật. Nếu gặp lỗi này:
- Chạy từ **Terminal.app** (hoặc iTerm) thay vì từ SSH / headless / một số terminal trong IDE.
- Đảm bảo đang login vào macOS với session có đồ họa (không phải `ssh ...` từ máy khác).
- Thử mở một cửa sổ Terminal mới và chạy lại: `cd ... && ./buddhist-go -g examples/address_management.bl`.

**Coordinates and style:**
- Widget coordinates `(x, y)` use **top-left as (0,0)**. Larger `y` = lower on the window.
- Optional colors: `gui_window` accepts `backgroundColor`; `gui_button` accepts `bgColor` and `textColor`. Each is a hash `{"r": 0.9, "g": 0.9, "b": 0.9}` with values 0–1 (or 0–255). Example: `gui_button(window, { "text": "OK", "x": 10, "y": 10, "width": 80, "height": 30, "bgColor": {"r": 0.2, "g": 0.5, "b": 0.9}, "onClick": fn() { } })`.
- Button text alignment: `gui_button` accepts optional `textAlign`: `"left"` (default), `"center"`, or `"right"`.

**Note:**
- Windows are created when `gui_run()` is called, not immediately.
- The event loop runs until all windows are closed.
- See `examples/gui_example.bl` for a basic example and `examples/address_management.bl` for a table example.

### HTTP Functions

| Function | Description |
|----------|-------------|
| `http_request(config)` | Make an HTTP request with config (url, method, headers, body, timeout_ms, progress callback) |
| `curl(config)` | Alias for `http_request` |

**Example:**

```javascript
place response = http_request({
    "url": "https://api.example.com/data",
    "method": "GET",
    "headers": {"Authorization": "Bearer token"},
    "timeout_ms": 5000
});

println(response["status"]);  // HTTP status code
println(response["body"]);    // Response body
println(response["headers"]); // Response headers
```

### File I/O Functions

| Function | Description |
|----------|-------------|
| `readFile(path)` | Read the entire contents of a file as a string |
| `writeFile(path, content)` | Write a string or blob to a file (creates file if it doesn't exist) |
| `readDir(path)` | Read directory contents and return an array of file/directory names |

**Example:**

```javascript
// Write a file
writeFile("test.txt", "Hello, World!\nThis is a test file.");

// Read the file back
place content = readFile("test.txt");
println(content);

// List directory contents
place files = readDir(".");
place i = 0;
while (i < len(files)) {
    println("  - " + files[i]);
    i = i + 1;
}
```

**Note:**
- `readFile` returns the file content as a string
- `writeFile` accepts either a string or blob as content
- `readDir` returns an array of strings (file/directory names), sorted alphabetically
- File paths are relative to the current working directory
- See `examples/file_io_example.bl` for a complete example

## Project Structure

```
buddhist-go/
├── cmd/
│   └── buddhist/        # CLI entry point
├── pkg/
│   ├── ast/             # Abstract Syntax Tree
│   ├── code/            # Bytecode opcodes and instructions
│   ├── compiler/        # AST to bytecode compiler
│   ├── lexer/           # Tokenizer (standard and optimized)
│   ├── object/          # Runtime object system
│   ├── parser/          # Token to AST parser
│   ├── token/           # Token definitions
│   ├── tracing/         # Debug tracing utilities
│   └── vm/              # Virtual machine (standard and optimized)
├── examples/            # Example scripts
└── intellij-plugin/     # IDE plugin for syntax highlighting
```

## Architecture

```
Source Code → Lexer → Tokens → Parser → AST → Compiler → Bytecode → VM → Result
```

1. **Lexer**: Converts source code into tokens
2. **Parser**: Builds an Abstract Syntax Tree from tokens
3. **Compiler**: Compiles AST to bytecode with optimizations
4. **VM**: Executes bytecode using a stack-based virtual machine

## Performance

The interpreter includes several performance optimizations:

- **Optimized VM**: Uses cached frame references and inline operations
- **Integer Caching**: Pre-allocates frequently used small integers
- **Constant Folding**: Evaluates constant expressions at compile time
- **Byte Slice Lexer**: Reduces string allocations during tokenization
- **Parallel Array Operations**: Uses goroutines for large array operations (>1000 elements)

### Benchmarking

```bash
./buddhist-go --benchmark examples/benchmark.bl
```

## Development

### Building

```bash
go build ./...
```

### Testing

```bash
go test ./...
```

### Running Examples

```bash
go run . examples/hello.bl
go run . examples/fizzbuzz.bl
go run . examples/gui_example.bl
go run . examples/http_request.bl
go run . examples/file_io_example.bl
```

## IDE Support

An IntelliJ/WebStorm plugin is available in the `intellij-plugin/` directory for syntax highlighting and basic language support.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Roadmap

### Short-term (1-2 weeks)
- ✅ Progress callbacks for HTTP requests
- ✅ Escape sequence support in strings (`\n`, `\t`, etc.)
- ✅ Math functions: `sqrt()`, `pow()`, `abs()`, `floor()`, `ceil()`, `round()`, `sin()`, `cos()`, `tan()`, `log()`, `exp()`, `random()`
- ✅ String functions: `split()`, `join()`, `trim()`, `substring()`, `replace()`, `upper()`, `lower()`, `startsWith()`, `endsWith()`, `repeat()`
- ✅ Array functions: `map()`, `filter()`, `reduce()`, `reverse()`, `contains()`, `indexOf()`, `unique()`, `flatten()`, `sum()`, `min()`, `max()`, `avg()`
- ✅ GUI functions: `gui_window()`, `gui_button()`, `gui_show()`, `gui_alert()`, `gui_close()`, `gui_run()` (powered by faiface/pixel)
- ✅ File I/O: `readFile()`, `writeFile()`, `readDir()`

### Medium-term (1 month)
- 🔲 Module/Import system: `import "utils.bl"`
- 🔲 Better error handling with stack traces
- 🔲 Array functions: `sort()`, `find()`
- 🔲 Date/Time functions: `now()`, `formatDate()`, `parseDate()`
- 🔲 Code formatter: `buddhist fmt`
- 🔲 Linter: `buddhist lint`

### Long-term (2-3 months)
- 🔲 Testing framework with built-in test runner
- 🔲 Package manager for dependency management
- 🔲 OOP support (classes and objects)
- 🔲 Type system (optional type hints)
- 🔲 Standard library with collections, networking, crypto
- 🔲 Profiler and performance analysis tools
- 🔲 Documentation generator

### IDE/Editor Enhancements
- 🔲 Code completion (IntelliSense)
- 🔲 Go to definition
- 🔲 Find usages
- 🔲 Refactoring support
- 🔲 Real-time error highlighting
- 🔲 Better REPL with history and auto-completion

## License

MIT License - See [LICENSE](LICENSE) for details.

## Acknowledgments

This interpreter is inspired by the book "Writing An Interpreter In Go" by Thorsten Ball, with additional features for concurrency, optimization, and PHP-style arrays.
