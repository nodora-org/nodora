package core

import (
	"encoding/json"
	"fmt"
)

type Value struct {
	Raw       any
	Undefined bool
}

func (v Value) String() string {
	if v.Undefined {
		return "undefined"
	}
	return fmt.Sprintf("%v", v.Raw)
}

func (v *Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ToRaw())
}

func (v *Value) UnmarshalJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	nv, err := NormalizeValue(raw)
	if err != nil {
		return err
	}
	*v = nv
	return nil
}

func fromRaw(raw any) Value {
	return Value{Raw: raw}
}

func V(raw any) Value {
	v, _ := NormalizeValue(raw)
	return v
}

func VP(raw any) *Value {
	v := V(raw)
	return &v
}

func U() Value {
	return Value{
		Raw:       nil,
		Undefined: true,
	}
}

func (v *Value) ToRaw() any {
	if v.Undefined {
		return nil
	}
	switch n := v.Raw.(type) {
	case Value:
		return n.ToRaw()
	case []Value:
		arr := make([]any, len(n))
		for i, v := range n {
			arr[i] = v.ToRaw()
		}
		return arr
	case ValueMap:
		raw := make(map[string]any, len(n))
		for k, v := range n {
			raw[k] = v.ToRaw()
		}
		return raw
	default:
		return n
	}
}

func (v *Value) Type() string {
	if v.Undefined {
		return "undefined"
	}
	switch inner := v.Raw.(type) {
	case Value:
		return inner.Type()
	case []Value:
		return "array"
	case ValueMap:
		return "object"
	default:
		return fmt.Sprintf("%T", v.Raw)
	}
}

type ValueMap map[string]Value

func NewValueMap(raw map[string]any) (ValueMap, error) {
	result := make(ValueMap, len(raw))
	for k, v := range raw {
		nv, err := NormalizeValue(v)
		if err != nil {
			return nil, err
		}
		result[k] = nv
	}
	return result, nil
}

func (n *ValueMap) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*n = make(ValueMap, len(raw))
	for k, v := range raw {
		nv, err := NormalizeValue(v)
		if err != nil {
			return err
		}
		(*n)[k] = nv
	}
	return nil
}

func NormalizeValue(v any) (Value, error) {
	switch val := v.(type) {
	case nil:
		return fromRaw(nil), nil
	case Value:
		return val, nil
	case []Value:
		return fromRaw(val), nil
	case ValueMap:
		return fromRaw(val), nil

	// primitives that can be stored as-is
	case string, bool, float64:
		return fromRaw(val), nil

	// numeric types -> convert to float64
	case float32:
		return fromRaw(float64(val)), nil
	case int:
		return fromRaw(float64(val)), nil
	case int8:
		return fromRaw(float64(val)), nil
	case int16:
		return fromRaw(float64(val)), nil
	case int32:
		return fromRaw(float64(val)), nil
	case int64:
		return fromRaw(float64(val)), nil
	case uint:
		return fromRaw(float64(val)), nil
	case uint8:
		return fromRaw(float64(val)), nil
	case uint16:
		return fromRaw(float64(val)), nil
	case uint32:
		return fromRaw(float64(val)), nil
	case uint64:
		return fromRaw(float64(val)), nil

	// mixed maps
	case map[string]any:
		nm, err := NewValueMap(val)
		if err != nil {
			return U(), err
		}
		return fromRaw(nm), nil

	// slices
	case []any:
		return normalizeSlice(val)
	case []string:
		return normalizeSlice(val)
	case []bool:
		return normalizeSlice(val)
	case []float64:
		return normalizeSlice(val)

	// other numeric slices - convert to []float64
	case []float32:
		return normalizeSlice(val)
	case []int:
		return normalizeSlice(val)
	case []int64:
		return normalizeSlice(val)
	case []int32:
		return normalizeSlice(val)
	case []int16:
		return normalizeSlice(val)
	case []int8:
		return normalizeSlice(val)
	case []uint:
		return normalizeSlice(val)
	case []uint64:
		return normalizeSlice(val)
	case []uint32:
		return normalizeSlice(val)
	case []uint16:
		return normalizeSlice(val)
	case []uint8:
		return normalizeSlice(val)

	default:
		return U(), fmt.Errorf("conversion failed on value %v of type %T", val, val)
	}
}

func normalizeSlice[T any](arr []T) (Value, error) {
	if len(arr) == 0 {
		return fromRaw([]Value{}), nil
	}
	result := make([]Value, len(arr))
	for i, item := range arr {
		nv, err := NormalizeValue(item)
		if err != nil {
			return U(), err
		}
		result[i] = nv
	}
	return fromRaw(result), nil
}
