package types

import (
	"testing"
)

func TestFormatSignature(t *testing.T) {
	fn := Func{
		Name: "len",
		Args: []ArgSpec{{
			Name:     "value",
			Type:     NewArrayType(AnyType),
			Required: true,
		}},
		ReturnType: NumberType,
		Fn:         nil,
	}
	sig := fn.Signature()
	expected := "len(value: array<any>) -> number"
	if sig != expected {
		t.Errorf("Expected signature '%s', got '%s'", expected, sig)
	}
}

func TestCompactSignature(t *testing.T) {
	fn := Func{
		Name: "len",
		Args: []ArgSpec{{
			Name:     "value",
			Type:     NewArrayType(AnyType),
			Required: true,
		}},
		ReturnType: NumberType,
		Fn:         nil,
	}
	sig := fn.CompactSignature()
	expected := "len(value)"
	if sig != expected {
		t.Errorf("Expected signature '%s', got '%s'", expected, sig)
	}
}
