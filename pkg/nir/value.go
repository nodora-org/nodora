package nir

import "encoding/json"

type Value any
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
