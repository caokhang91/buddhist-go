package tracing

import (
	"fmt"
	"os"
	"time"
)

var enabled = false // Disabled by default for performance

// Enable enables tracing
func Enable() {
	enabled = true
}

// Disable disables tracing
func Disable() {
	enabled = false
}

// IsEnabled returns whether tracing is enabled
func IsEnabled() bool {
	return enabled
}

// Trace prints a trace message if tracing is enabled
func Trace(format string, args ...interface{}) {
	if enabled {
		fmt.Fprintf(os.Stderr, "[TRACE] "+format+"\n", args...)
	}
}

// TraceNetwork prints a network I/O trace message
func TraceNetwork(format string, args ...interface{}) {
	if enabled {
		fmt.Fprintf(os.Stderr, "[NETWORK] "+format+"\n", args...)
	}
}

// TraceCPU prints a CPU-bound operation trace message
func TraceCPU(format string, args ...interface{}) {
	if enabled {
		fmt.Fprintf(os.Stderr, "[CPU] "+format+"\n", args...)
	}
}

// TraceTiming prints a timing trace message
func TraceTiming(operation string, duration time.Duration) {
	if enabled {
		fmt.Fprintf(os.Stderr, "[TIMING] %s took %v\n", operation, duration)
	}
}

// TraceStart marks the start of an operation
func TraceStart(operation string) func() {
	if !enabled {
		return func() {}
	}
	start := time.Now()
	TraceCPU("Starting: %s", operation)
	return func() {
		duration := time.Since(start)
		TraceTiming(operation, duration)
		TraceCPU("Completed: %s", operation)
	}
}
