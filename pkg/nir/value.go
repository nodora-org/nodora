package nir

import (
	"encoding/json"
	"fmt"
)

type Value struct {
	Raw       any
	Undefined bool
}

func (v *Value) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Raw)
}

func (v *Value) UnmarshalJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	v.Raw = raw
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
	switch v.Raw.(type) {
	case []Value:
		return "array"
	case map[string]Value:
		return "object"
	default:
		return fmt.Sprintf("%T", v.Raw)
	}
}

type ValueMap map[string]Value

func NewValueMap(raw map[string]any) ValueMap {
	result := make(map[string]Value, len(raw))
	for k, v := range raw {
		result[k] = normalizeValue(v)
	}
	return result
}

func (n *ValueMap) UnmarshalJSON(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*n = make(map[string]Value, len(raw))
	for k, v := range raw {
		(*n)[k] = normalizeValue(v)
	}
	return nil
}
