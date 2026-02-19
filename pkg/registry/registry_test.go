package registry_test

import (
	"testing"

	"nodora.org/nodora/pkg/registry"
)

func TestRegistration(t *testing.T) {
	fn, exists := registry.Global().Get("", "is_defined")
	if !exists {
		t.Error("Expected 'is_defined' function to be registered in 'core' namespace")
	}
	if fn.Name != "is_defined" {
		t.Errorf("Expected function name 'is_defined', got '%s'", fn.Name)
	}
	if fn.Namespace != "" {
		t.Errorf("Expected empty namespace', got '%s'", fn.Namespace)
	}
}
