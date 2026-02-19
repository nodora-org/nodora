package core

func IntPtr(i int) *int {
	return &i
}

func IsFloat(v any) bool {
	switch n := v.(type) {
	case Value:
		return IsFloat(n.Raw)
	case float32, float64:
		return true
	default:
		return false
	}
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

func ToInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case Value:
		return ToInt64(n.Raw)
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

func ToInt(v any) (int, bool) {
	switch n := v.(type) {
	case Value:
		return ToInt(n.Raw)
	case int:
		return n, true
	case int64:
		return int(n), true
	case float32:
		return int(n), true
	case float64:
		return int(n), true
	}
	return 0, false
}
