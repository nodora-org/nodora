package core

import (
	"encoding/json"
	"fmt"
)

type Value struct {
	Raw       any
	Undefined bool
}

func (v *Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.ToRaw())
}

func (v *Value) UnmarshalJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*v = NormalizeValue(raw)
	return nil
}

func V(raw any) Value {
	return Value{Raw: raw}
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

func NewValueMap(raw map[string]any) ValueMap {
	result := make(ValueMap, len(raw))
	for k, v := range raw {
		result[k] = NormalizeValue(v)
	}
	return result
}

func NormalizeValue(v any) Value {
	switch val := v.(type) {
	case []any:
		result := make([]Value, len(val))
		for i, item := range val {
			result[i] = NormalizeValue(item)
		}
		return V(result)
	case map[string]any:
		return V(NewValueMap(val))
	default:
		// primitives (string, float64, bool, nil) stay as-is
		return V(val)
	}
}

func (n *ValueMap) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*n = make(ValueMap, len(raw))
	for k, v := range raw {
		(*n)[k] = NormalizeValue(v)
	}
	return nil
}
