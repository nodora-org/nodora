package tests

import (
	"strings"
	"testing"

	"nodora.org/nodora/pkg/compiler"
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/evaluator"
)

func TestDefinedFunction(t *testing.T) {
	source := `
signal TestSignal(msg)

rule TestRule {
    has_name = is_defined(input.name)
    has_age = is_defined(input.age)
    
    out result = has_name && has_age
    
    emit TestSignal("missing fields") when !result
}
`
	c := compiler.NewCompiler()
	program, err := c.Compile(source)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	e := evaluator.NewEvaluator(program)

	// Test with both fields present
	result1, err := e.EvaluateRule("TestRule", core.ValueMap{"name": core.V("John"), "age": core.V(30)})
	if err != nil {
		t.Fatalf("Failed to evaluate with both fields: %v", err)
	}
	if result1 == nil {
		t.Fatal("Expected result, got nil")
	}
	if result1.Outputs["result"] != true {
		t.Errorf("Expected result=true, got %v", result1.Outputs["result"])
	}

	// Test with missing field
	result2, err := e.EvaluateRule("TestRule", core.ValueMap{"name": core.V("John")})
	if err != nil {
		t.Fatalf("Failed to evaluate with missing field: %v", err)
	}
	if result2 == nil {
		t.Fatal("Expected result, got nil")
	}
	if result2.Outputs["result"] != false {
		t.Errorf("Expected result=false, got %v", result2.Outputs["result"])
	}
}

func TestUndefinedFunctionError(t *testing.T) {
	source := `
rule TestRule {
    out result = undefined_func(input.name)
}
`
	c := compiler.NewCompiler()
	_, err := c.Compile(source)
	if err == nil {
		t.Fatal("Expected error for undefined function, got nil")
	}
	if !strings.Contains(err.Error(), "undefined function 'undefined_func'") {
		t.Errorf("Expected 'undefined function' error, got: %v", err)
	}
}

func TestUndefinedNamespaceFunction(t *testing.T) {
	source := `
rule TestRule {
    out result = crypto::dummy("asd")
}
`
	c := compiler.NewCompiler()
	_, err := c.Compile(source)
	if err == nil {
		t.Fatal("Expected error for undefined namespace function, got nil")
	}
	if !strings.Contains(err.Error(), "undefined function 'crypto::dummy'") {
		t.Errorf("Expected 'undefined function' error with namespace, got: %v", err)
	}
}
