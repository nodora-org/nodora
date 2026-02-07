package tests

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nodora.org/nodora/internal/evaluator"
	"nodora.org/nodora/internal/nir"
	"nodora.org/nodora/internal/parser"
	"nodora.org/nodora/internal/semantics"
)

type TestSample struct {
	Input    nir.ValueMap               `json:"input"`
	Expected evaluator.EvaluationResult `json:"expected"`
}

func TestE2E(t *testing.T) {
	testRoot := "./samples"

	// Find all test directories
	err := filepath.Walk(testRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".rule") {
			return nil
		}

		testName := strings.TrimSuffix(filepath.Base(path), ".rule")
		testDir := filepath.Dir(path)
		inputsPath := filepath.Join(testDir, testName+".inputs.json")

		if _, err := os.Stat(inputsPath); os.IsNotExist(err) {
			t.Logf("Skipping %s: no inputs file found at %s", testName, inputsPath)
			return nil
		}

		t.Run(testName, func(t *testing.T) {
			runTest(t, path, inputsPath, testName)
		})

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk test directory: %v", err)
	}
}

func runTest(t *testing.T, rulePath, inputsPath, testName string) {
	ruleContent, err := os.ReadFile(rulePath)
	if err != nil {
		t.Fatalf("Failed to read rule file %s: %v", rulePath, err)
	}

	ast, err := parser.Parse(string(ruleContent))
	if err != nil {
		t.Fatalf("Failed to parse rule %s: %v", testName, err)
	}

	analyzer := semantics.NewSemanticAnalyzer()
	errors := analyzer.Analyze(ast)
	if len(errors) > 0 {
		for _, e := range errors {
			t.Errorf("Semantic error in %s: %v", testName, e)
		}
		t.FailNow()
	}

	converter := nir.NewConverter()
	program, err := converter.ConvertFromAST(ast)
	if err != nil {
		t.Fatalf("Failed to convert rule %s: %v", testName, err)
	}

	ev := evaluator.NewEvaluator(&program)

	inputsContent, err := os.ReadFile(inputsPath)
	if err != nil {
		t.Fatalf("Failed to read inputs file %s: %v", inputsPath, err)
	}

	var testSamples []TestSample
	if err := json.Unmarshal(inputsContent, &testSamples); err != nil {
		t.Fatalf("Failed to parse inputs file %s: %v", inputsPath, err)
	}

	ruleNames := ev.GetRuleNames()
	if len(ruleNames) == 0 {
		t.Fatalf("No rules found in %s", testName)
	}

	for i, tc := range testSamples {
		t.Run(fmt.Sprintf("sample_%d", i), func(t *testing.T) {
			result, err := ev.EvaluateRule(ruleNames[0], tc.Input)
			if err != nil {
				t.Fatalf("Evaluation error in test %d: %v", i, err)
			}

			for key, expectedVal := range tc.Expected.Outputs {
				actualVal, exists := result.Outputs[key]
				if !exists {
					t.Errorf("Test %d: expected output '%s' not found", i, key)
					continue
				}

				if !compareValues(expectedVal, actualVal) {
					t.Errorf("Test %d: output '%s' mismatch:\n  expected: %v (%T)\n  actual:   %v (%T)",
						i, key, expectedVal, expectedVal, actualVal, actualVal)
				}
			}

			for key := range result.Outputs {
				if _, exists := tc.Expected.Outputs[key]; !exists {
					t.Errorf("Test %d: unexpected output '%s' with value %v", i, key, result.Outputs[key])
				}
			}

			expectedSignals := make([]nir.EmittedSignal, 0)
			if tc.Expected.Signals != nil {
				expectedSignals = tc.Expected.Signals
			}

			if !compareValues(result.Signals, expectedSignals) {
				t.Errorf("Test %d: mismatch in emitted signals (result: %v, expected: %v)", i, result.Signals, expectedSignals)
			}
		})
	}
}

const EPSILON = 1e-6

func floatEquals(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}
func compareValues(expected, actual any) bool {
	switch exp := expected.(type) {
	case float64:
		if act, ok := actual.(float64); ok {
			return floatEquals(exp, act, EPSILON)
		}
	}
	expectedJSON, err1 := json.Marshal(expected)
	actualJSON, err2 := json.Marshal(actual)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(expectedJSON) == string(actualJSON)
}
