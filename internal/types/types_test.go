package types

import (
	"testing"
)

func TestTypeSystem(t *testing.T) {
	t.Run("SimpleType equality", func(t *testing.T) {
		if !StringType.Equals(StringType) {
			t.Error("StringType should equal itself")
		}
		if StringType.Equals(NumberType) {
			t.Error("StringType should not equal NumberType")
		}
	})

	t.Run("ArrayType equality", func(t *testing.T) {
		arr1 := NewArrayType(NumberType)
		arr2 := NewArrayType(NumberType)
		arr3 := NewArrayType(StringType)

		if !arr1.Equals(arr2) {
			t.Error("array<number> should equal array<number>")
		}
		if arr1.Equals(arr3) {
			t.Error("array<number> should not equal array<string>")
		}
	})

	t.Run("IsAssignableFrom", func(t *testing.T) {
		// Any type accepts any value
		if !AnyType.IsAssignableFrom(StringType) {
			t.Error("any should be assignable from string")
		}
		if !AnyType.IsAssignableFrom(NumberType) {
			t.Error("any should be assignable from number")
		}

		// Unknown is assignable to any type
		if !StringType.IsAssignableFrom(UnknownType) {
			t.Error("string should accept unknown")
		}

		// Exact matches
		if !StringType.IsAssignableFrom(StringType) {
			t.Error("string should accept string")
		}
		if StringType.IsAssignableFrom(NumberType) {
			t.Error("string should not accept number")
		}
	})

	t.Run("IsArray helper", func(t *testing.T) {
		if !IsArray(NewArrayType(StringType)) {
			t.Error("IsArray should return true for array types")
		}
		if IsArray(StringType) {
			t.Error("IsArray should return false for simple types")
		}
	})

	t.Run("GetArrayElement", func(t *testing.T) {
		arr := NewArrayType(NumberType)
		elem := GetArrayElement(arr)
		if elem == nil {
			t.Error("GetArrayElement should return element type for arrays")
		}
		if !elem.Equals(NumberType) {
			t.Error("GetArrayElement should return the correct element type")
		}
	})
}
