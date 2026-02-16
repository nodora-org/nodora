package nir

import (
	"fmt"
	"slices"
)

type Float interface {
	~float32 | ~float64
}

type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Numeric interface {
	Signed | Unsigned | Float
}

func IntPtr(i int) *int {
	return &i
}

func isFloat(v any) bool {
	switch v.(type) {
	case float32, float64:
		return true
	default:
		return false
	}
}

func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	default:
		return 0, false
	}
}

func toInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int8:
		return int64(n), true
	case int16:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case uint:
		return int64(n), true
	case uint8:
		return int64(n), true
	case uint16:
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint64:
		return int64(n), true
	default:
		return 0, false
	}
}

func toInt(v any) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float32:
		return int(val), true
	case float64:
		return int(val), true
	}
	return 0, false
}

func normalizeValue(v any) Value {
	switch val := v.(type) {
	case []any:
		// convert []any to []Value
		result := make([]Value, len(val))
		for i, item := range val {
			result[i] = normalizeValue(item)
		}
		return result
	case map[string]any:
		// convert map[string]any to map[string]Value
		result := make(ValueMap, len(val))
		for k, item := range val {
			result[k] = normalizeValue(item)
		}
		return result
	default:
		// primitives (string, float64, bool, nil) stay as-is
		return val
	}
}

func containsValue(slice, value any) (bool, error) {
	switch s := slice.(type) {
	case []Value:
		return slices.Contains(s, value), nil
	case []any:
		return slices.Contains(s, value), nil
	default:
		return false, fmt.Errorf("expected array, got %T", slice)
	}
}

func validateArgs(o *Op, argCount int) error {
	if len(o.Args) != argCount {
		return fmt.Errorf("invalid operation arguments for '%s'", o.Kind)
	}
	return nil
}

func validateOp(o *Op, ctx *EvaluationContext, argCount int) error {
	if err := validateArgs(o, argCount); err != nil {
		return err
	}
	if o.Out == nil {
		return fmt.Errorf("output index required for '%s'", o.Kind)
	}
	if *o.Out < 0 || *o.Out >= len(ctx.Slots) {
		return fmt.Errorf("output index %d is out of bounds", *o.Out)
	}
	return nil
}
