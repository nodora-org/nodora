package core

import "math"

func (v Value) AsFloat() (float64, bool) {
	if f, ok := v.Raw.(float64); ok {
		return f, true
	}
	return ToFloat64(v.Raw)
}

func (v Value) AsInt() (int, bool) {
	f, ok := v.AsFloat()
	if !ok {
		return 0, false
	}
	return int(f), true
}

func (v Value) AsIndex() (int, bool) {
	f, ok := v.AsFloat()
	if !ok || f != math.Trunc(f) {
		return 0, false
	}
	return int(f), true
}

func (v Value) AsString() (string, bool) {
	s, ok := v.Raw.(string)
	return s, ok
}

func (v Value) AsBool() (bool, bool) {
	b, ok := v.Raw.(bool)
	return b, ok
}

func (v Value) AsArray() ([]Value, bool) {
	a, ok := v.Raw.([]Value)
	return a, ok
}

func (v Value) AsObject() (ValueMap, bool) {
	m, ok := v.Raw.(ValueMap)
	return m, ok
}

func (v Value) AsLambda() (*Lambda, bool) {
	l, ok := v.Raw.(*Lambda)
	return l, ok
}

func (v Value) IsUndefined() bool { return v.Undefined }
func (v Value) IsNull() bool      { return !v.Undefined && v.Raw == nil }
