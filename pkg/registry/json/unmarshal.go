package json

import (
	"encoding/json"
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func unmarshal() types.Func {
	return types.Func{
		Name:       "unmarshal",
		Args:       []types.ArgSpec{{Name: "str", Type: types.StringType, Required: true}},
		ReturnType: types.AnyType,
		Fn:         unmarshalImpl,
	}
}

func unmarshalImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	var raw any
	err := json.Unmarshal([]byte(strVal), &raw)
	if err != nil {
		return core.U(), err
	}
	return core.NormalizeValue(raw)
}
