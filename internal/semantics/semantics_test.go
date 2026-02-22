package semantics

import (
	"strings"
	"testing"

	"nodora.org/nodora/internal/parser"
)

func TestSemanticErrorsWithSpan(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedErrMsg string
	}{
		{
			name: "undefined variable",
			input: `
rule test {
    x = undefined_var + 1
}
`,
			expectedErrMsg: "3:9: undefined symbol 'undefined_var'\n    x = undefined_var + 1\n        ^",
		},
		{
			name: "type mismatch",
			input: `
rule test {
    x = "hello" + 5
}
`,
			expectedErrMsg: "3:9: operator '+' cannot be applied to 'string' and 'number'\n    x = \"hello\" + 5\n        ^",
		},
		{
			name: "undefined signal",
			input: `
rule test {
    emit undefined_signal()
}
`,
			expectedErrMsg: "3:5: undefined signal 'undefined_signal'\n    emit undefined_signal()\n    ^",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			analyzer := NewSemanticAnalyzer(tt.input)
			err = analyzer.Analyze(program)

			if semErrs, ok := err.(*SemanticErrors); ok {
				if semErrs.Count() == 0 {
					t.Errorf("Expected error for %s", tt.name)
					return
				}

				errMsg := semErrs.Errors[0].Error()
				if !strings.Contains(errMsg, ":") {
					t.Error("Expected error to contain position information (line:col)")
				}

				if errMsg != tt.expectedErrMsg {
					t.Errorf("Expected error to be '%s', but got '%s'", tt.expectedErrMsg, errMsg)
				}
			} else {
				t.Errorf("Unexpected error %s", err)
			}
		})
	}
}

func TestArrayTypeInference(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedType string
		shouldError  bool
	}{
		{
			name: "homogeneous number array",
			input: `
rule test {
    x = [1, 2, 3]
}
`,
			expectedType: "array<number>",
			shouldError:  false,
		},
		{
			name: "homogeneous string array",
			input: `
rule test {
    x = ["a", "b", "c"]
}
`,
			expectedType: "array<string>",
			shouldError:  false,
		},
		{
			name: "homogeneous bool array",
			input: `
rule test {
    x = [true, false]
}
`,
			expectedType: "array<bool>",
			shouldError:  false,
		},
		{
			name: "mixed array falls back to any",
			input: `
rule test {
    x = [1, "hello"]
}
`,
			expectedType: "array<any>",
			shouldError:  false,
		},
		{
			name: "empty array is array<any>",
			input: `
rule test {
    x = []
}
`,
			expectedType: "array<any>",
			shouldError:  false,
		},
		{
			name: "in operator requires array",
			input: `
rule test {
    x = 1 in "not an array"
}
`,
			expectedType: "",
			shouldError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			analyzer := NewSemanticAnalyzer(tt.input)
			err = analyzer.Analyze(program)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s", tt.name)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
		})
	}
}

func TestInputObjectType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name: "input access returns unknown",
			input: `
rule test {
    x = input.someField
}
`,
			shouldError: false,
		},
		{
			name: "cannot access property on non-object",
			input: `
rule test {
    x = "string".field
}
`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			analyzer := NewSemanticAnalyzer(tt.input)
			err = analyzer.Analyze(program)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTypeCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name: "unknown is compatible with any type",
			input: `
rule test {
    x = input.a + 5
}
`,
			shouldError: false,
		},
		{
			name: "string concatenation",
			input: `
rule test {
    x = "hello" + " world"
}
`,
			shouldError: false,
		},
		{
			name: "number arithmetic",
			input: `
rule test {
    x = 5 + 3 * 2
}
`,
			shouldError: false,
		},
		{
			name: "bool operations",
			input: `
rule test {
    x = true && false || true
}
`,
			shouldError: false,
		},
		{
			name: "comparison requires compatible types",
			input: `
rule test {
    x = "hello" == 5
}
`,
			shouldError: true,
		},
		{
			name: "arithmetic requires numbers",
			input: `
rule test {
    x = "hello" - 5
}
`,
			shouldError: true,
		},
		{
			name: "logical requires bools",
			input: `
rule test {
    x = true && 5
}
`,
			shouldError: true,
		},
		{
			name: "conditional requires bool condition",
			input: `
rule test {
    x = if "hello" then 1 else 2
}
`,
			shouldError: true,
		},
		{
			name: "conditional requires compatible branches",
			input: `
rule test {
    x = if true then 1 else "hello"
}
`,
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			analyzer := NewSemanticAnalyzer(tt.input)
			err = analyzer.Analyze(program)

			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for %s", tt.name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
