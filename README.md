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
>>> place x = 10
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
- ✅ Math functions: `sqrt()`, `pow()`, `abs()`, `floor()`, `ceil()`, `round()`, `sin()`, `cos()`, `tan()`, `log()`, `exp()`, `random()`
- ✅ String functions: `split()`, `join()`, `trim()`, `substring()`, `replace()`, `upper()`, `lower()`, `startsWith()`, `endsWith()`, `repeat()`
- ✅ Array functions: `map()`, `filter()`, `reduce()`, `reverse()`, `contains()`, `indexOf()`, `unique()`, `flatten()`, `sum()`, `min()`, `max()`, `avg()`
- 🔲 File I/O: `readFile()`, `writeFile()`, `readDir()`

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
