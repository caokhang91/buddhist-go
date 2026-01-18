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

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/caokhang91/buddhist-go.git
cd buddhist-go

# Build the interpreter
go build -o buddhist ./cmd/buddhist

# Or install directly
go install ./cmd/buddhist
```

### From Release

Download pre-built binaries from the [Releases](https://github.com/caokhang91/buddhist-go/releases) page.

## Quick Start

### Interactive REPL

```bash
./buddhist
```

```
╔══════════════════════════════════════════════════════════════╗
║               Go-Powered Interpreter Language                ║
║                     Version 1.0.0                            ║
╚══════════════════════════════════════════════════════════════╝

>>> println("Hello, World!")
Hello, World!
>>> let x = 10
>>> x * 2
20
```

### Run a Script

```bash
./buddhist examples/hello.bl
```

## Language Syntax

### Variables

```javascript
let x = 5;
const PI = 3.14159;
x = x + 1;
```

### Functions

```javascript
fn add(a, b) {
    return a + b;
}

// Anonymous functions
let multiply = fn(a, b) { a * b };

// Closures
fn counter() {
    let count = 0;
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
for (let i = 0; i < 10; i = i + 1) {
    println(i);
}
```

### Arrays

```javascript
// Standard arrays
let arr = [1, 2, 3, 4, 5];
println(arr[0]);  // 1
println(len(arr)); // 5

// PHP-style arrays with keys
let map = [
    "name" => "Buddhist",
    "version" => "1.0.0",
    0 => "indexed"
];
println(map["name"]);  // Buddhist
```

### Hash Maps

```javascript
let person = {
    "name": "John",
    "age": 30
};
println(person["name"]);  // John
```

### Concurrency

```javascript
// Create a channel
let ch = channel;

// Spawn a goroutine
spawn fn() {
    ch <- "Hello from goroutine!";
};

// Receive from channel
let msg = <-ch;
println(msg);
```

## Built-in Functions

| Function | Description |
|----------|-------------|
| `println(...)` | Print values with newline |
| `print(...)` | Print values without newline |
| `len(x)` | Get length of array/string |
| `first(arr)` | Get first element of array |
| `last(arr)` | Get last element of array |
| `rest(arr)` | Get array without first element |
| `push(arr, val)` | Append value to array |
| `type(x)` | Get type of value |
| `str(x)` | Convert to string |
| `int(x)` | Convert to integer |
| `float(x)` | Convert to float |

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
./buddhist --benchmark examples/benchmark.bl
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
go run ./cmd/buddhist examples/hello.bl
```

## IDE Support

An IntelliJ/WebStorm plugin is available in the `intellij-plugin/` directory for syntax highlighting and basic language support.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Roadmap

### Short-term (1-2 weeks)
- ✅ Progress callbacks for HTTP requests
- ✅ Escape sequence support in strings (`\n`, `\t`, etc.)
- 🔲 Math functions: `sqrt()`, `pow()`, `abs()`, `floor()`, `ceil()`, `round()`
- 🔲 String functions: `split()`, `join()`, `trim()`, `substring()`, `indexOf()`, `replace()`
- 🔲 File I/O: `readFile()`, `writeFile()`, `readDir()`

### Medium-term (1 month)
- 🔲 Module/Import system: `import "utils.bl"`
- 🔲 Better error handling with stack traces
- 🔲 Array functions: `map()`, `filter()`, `reduce()`, `sort()`, `reverse()`, `find()`
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
