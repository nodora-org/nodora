package registry_test

import (
	"testing"

	"nodora.org/nodora/pkg/registry"
)

func TestRegistration(t *testing.T) {
	fn, exists := registry.Get("", "is_defined")
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

func TestFullPath(t *testing.T) {
	fn, _ := registry.Get("", "is_defined")
	if fn.FullPath() != "::is_defined" {
		t.Errorf("Unexpected FullName '%s'", fn.FullPath())
	}
}

func TestFormatSignature(t *testing.T) {
	fn, _ := registry.Get("", "is_defined")
	sig := fn.Signature()
	expected := "::is_defined(value) -> bool"
	if sig != expected {
		t.Errorf("Expected signature '%s', got '%s'", expected, sig)
	}
}

func TestCompactSignature(t *testing.T) {
	fn, _ := registry.Get("", "len")
	sig := fn.CompactSignature()
	expected := "::len(value)"
	if sig != expected {
		t.Errorf("Expected signature '%s', got '%s'", expected, sig)
	}
}
