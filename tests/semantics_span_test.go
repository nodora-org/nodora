package tests

import (
	"strings"
	"testing"

	"nodora.org/nodora/internal/parser"
	"nodora.org/nodora/internal/semantics"
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
			expectedErrMsg: "3:8: undefined symbol 'undefined_var'",
		},
		{
			name: "type mismatch",
			input: `
rule test {
    x = "hello" + 5
}
`,
			expectedErrMsg: "3:8: operator '+' cannot be applied to string and number",
		},
		{
			name: "undefined signal",
			input: `
rule test {
    emit undefined_signal()
}
`,
			expectedErrMsg: "3:4: undefined signal 'undefined_signal'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			analyzer := semantics.NewSemanticAnalyzer()
			errors := analyzer.Analyze(program)

			if len(errors) == 0 {
				t.Errorf("Expected error for %s", tt.name)
				return
			}

			errMsg := errors[0].Error()
			if !strings.Contains(errMsg, ":") {
				t.Error("Expected error to contain position information (line:col)")
			}

			if errMsg != tt.expectedErrMsg {
				t.Errorf("Expected error to be '%s', but got: %s", tt.expectedErrMsg, errMsg)
			}
		})
	}
}
