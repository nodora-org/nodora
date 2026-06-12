package semantics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nodora.org/nodora/internal/parser"
	_ "nodora.org/nodora/pkg/registry/all"
)

func TestSemantics(t *testing.T) {
	files, _ := filepath.Glob("testdata/*.rule")

	for _, file := range files {
		testName := strings.TrimSuffix(filepath.Base(file), ".rule")
		t.Run(testName, func(t *testing.T) {
			srcBytes, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}

			src := string(srcBytes)
			expected := extractExpectedErrors(src)
			src = stripErrorComments(src)

			program, err := parser.Parse(src)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			analyzer := NewSemanticAnalyzer(src)
			err = analyzer.Analyze(program)

			if err == nil && len(expected) > 0 {
				t.Fatalf("Expected %d error(s) but got none", len(expected))
			}

			if err != nil {
				if semErrs, ok := err.(*SemanticErrors); ok {
					compareErrors(t, expected, semErrs)
				}
			}
		})
	}
}

type ExpectedError struct {
	Line    int
	Message string
}

func extractExpectedErrors(src string) []ExpectedError {
	var expected []ExpectedError

	lines := strings.Split(src, "\n")
	for i, line := range lines {
		_, after, ok := strings.Cut(line, "// ERROR:")
		if !ok {
			continue
		}

		msg := strings.TrimSpace(after)
		expected = append(expected, ExpectedError{
			Line:    i + 1,
			Message: msg,
		})
	}

	return expected
}

func stripErrorComments(src string) string {
	var cleaned []string

	lines := strings.Split(src, "\n")
	for _, line := range lines {
		if idx := strings.Index(line, "// ERROR:"); idx != -1 {
			line = line[:idx]
		}
		cleaned = append(cleaned, line)
	}

	return strings.Join(cleaned, "\n")
}

func compareErrors(t *testing.T, expected []ExpectedError, actual *SemanticErrors) {
	if len(expected) != actual.Count() {
		t.Fatalf("Expected %d error(s), got %d:\n%s",
			len(expected),
			actual.Count(),
			actual.Error(),
		)
	}

	unmatched := make([]string, 0, len(actual.Errors))
	for _, e := range actual.Errors {
		unmatched = append(unmatched, e.Error())
	}

	for _, exp := range expected {
		found := false

		for i, msg := range unmatched {
			if strings.Contains(msg, exp.Message) {
				unmatched = append(unmatched[:i], unmatched[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Expected error on line %d: %q", exp.Line, exp.Message)
		}
	}
}
