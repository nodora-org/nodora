package core

import (
	"math"
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

func ToFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case Value:
		return ToFloat64(n.Raw)
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

func IsInt(v any) bool {
	switch n := v.(type) {
	case Value:
		return IsInt(n.Raw)
	case int, int8, int16, int32, int64:
		return true
	case float32:
		return float64(n) == math.Trunc(float64(n))
	case float64:
		return n == math.Trunc(n)
	default:
		return false
	}
}

func ToInt(v any) (int, bool) {
	switch n := v.(type) {
	case Value:
		return ToInt(n.Raw)
	case int:
		return n, true
	case int8:
		return int(n), true
	case int16:
		return int(n), true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float32:
		return int(n), true
	case float64:
		return int(n), true
	}
	return 0, false
}

const epsilon = 1e-9

func floatEquals(a, b float64) bool {
	diff := math.Abs(a - b)
	if diff <= epsilon {
		return true
	}
	return diff <= epsilon*math.Max(math.Abs(a), math.Abs(b))
}

func SafeEquals(a, b Value) bool {
	switch av := a.Raw.(type) {
	case nil:
		return b.Raw == nil
	case string:
		bv, ok := b.Raw.(string)
		return ok && av == bv
	case bool:
		bv, ok := b.Raw.(bool)
		return ok && av == bv
	case float64:
		bv, ok := b.Raw.(float64)
		return ok && floatEquals(av, bv)
	case Value:
		bv, ok := b.Raw.(Value)
		return ok && SafeEquals(av, bv)
	case []Value:
		bv, ok := b.Raw.([]Value)
		if !ok {
			return false
		}
		if &av == &bv {
			return true
		}
		if len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !SafeEquals(av[i], bv[i]) {
				return false
			}
		}
		return true
	case ValueMap:
		bv, ok := b.Raw.(ValueMap)
		if !ok {
			return false
		}
		if &av == &bv {
			return true
		}
		for k, v := range av {
			bvv, exists := bv[k]
			if !exists || !SafeEquals(v, bvv) {
				return false
			}
		}
		return true
	}
	return false
}
