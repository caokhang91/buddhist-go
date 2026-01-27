package vm

import (
	"testing"
)

func TestTryCatchThrow(t *testing.T) {
	input := `
	place x = 0;
	try {
		throw "boom";
		x = 1;
	} catch (e) {
		x = 2;
	}
	x;
	`

	evaluated := testEval(input)
	testObject(t, evaluated, int64(2))
}

func TestTryCatchNoThrow(t *testing.T) {
	input := `
	place x = 0;
	try {
		x = 1;
	} catch (e) {
		x = 2;
	}
	x;
	`

	evaluated := testEval(input)
	testObject(t, evaluated, int64(1))
}

func TestCatchBindsThrownValue(t *testing.T) {
	input := `
	place x = 0;
	try {
		throw 5;
	} catch (e) {
		x = e;
	}
	x;
	`

	evaluated := testEval(input)
	testObject(t, evaluated, int64(5))
}

func TestFinallyRunsOnNormalAndThrow(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name: "normal",
			input: `
			place x = 0;
			try {
				x = 1;
			} catch (e) {
				x = 2;
			} finally {
				x = x + 10;
			}
			x;
			`,
			expected: int64(11),
		},
		{
			name: "throw",
			input: `
			place x = 0;
			try {
				throw "boom";
			} catch (e) {
				x = 2;
			} finally {
				x = x + 10;
			}
			x;
			`,
			expected: int64(12),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evaluated := testEval(tt.input)
			testObject(t, evaluated, tt.expected)
		})
	}
}

// TestBuiltinErrorAsThrow ensures builtin functions that return error objects
// are treated as throws so try/catch can catch them.
func TestBuiltinErrorAsThrow(t *testing.T) {
	input := `
	place x = 0;
	try {
		place n = int("not a number");
		x = 1;
	} catch (e) {
		x = 2;
	}
	x;
	`
	evaluated := testEval(input)
	testObject(t, evaluated, int64(2))
}

