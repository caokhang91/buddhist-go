package object

import (
	"os"
	"sort"

	"github.com/caokhang91/buddhist-go/pkg/tracing"
)

// readFileBuiltin reads the entire contents of a file as a string
// Usage: readFile("path/to/file.txt")
func readFileBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	path, errObj := stringArg(args, 0, "readFile")
	if errObj != nil {
		return errObj
	}

	if tracing.IsEnabled() {
		tracing.Trace("Reading file: %s", path)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return newError("readFile failed: %s", err.Error())
	}

	if tracing.IsEnabled() {
		tracing.Trace("Read file: %s (%d bytes)", path, len(content))
	}

	return &String{Value: string(content)}
}

// writeFileBuiltin writes a string to a file
// Usage: writeFile("path/to/file.txt", "content")
func writeFileBuiltin(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}

	path, errObj := stringArg(args, 0, "writeFile")
	if errObj != nil {
		return errObj
	}

	var content []byte
	switch arg := args[1].(type) {
	case *String:
		content = []byte(arg.Value)
	case *Blob:
		content = arg.Data
	default:
		return newError("writeFile content must be STRING or BLOB, got %s", args[1].Type())
	}

	if tracing.IsEnabled() {
		tracing.Trace("Writing file: %s (%d bytes)", path, len(content))
	}

	err := os.WriteFile(path, content, 0644)
	if err != nil {
		return newError("writeFile failed: %s", err.Error())
	}

	if tracing.IsEnabled() {
		tracing.Trace("Wrote file: %s", path)
	}

	return &Null{}
}

// readDirBuiltin reads the contents of a directory and returns an array of file/directory names
// Usage: readDir("path/to/directory")
func readDirBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}

	path, errObj := stringArg(args, 0, "readDir")
	if errObj != nil {
		return errObj
	}

	if tracing.IsEnabled() {
		tracing.Trace("Reading directory: %s", path)
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return newError("readDir failed: %s", err.Error())
	}

	// Sort entries by name for consistent output
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	// Convert to array of strings
	elements := make([]Object, len(entries))
	for i, entry := range entries {
		elements[i] = &String{Value: entry.Name()}
	}

	if tracing.IsEnabled() {
		tracing.Trace("Read directory: %s (%d entries)", path, len(entries))
	}

	return &Array{Elements: elements}
}
