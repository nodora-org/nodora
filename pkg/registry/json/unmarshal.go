package json

import (
	"encoding/json"
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func unmarshal() types.Func {
	return types.Func{
		Name:        "unmarshal",
		Description: "Parses a JSON string into a value.",
		Args:        []types.ArgSpec{{Name: "str", Description: "The JSON string to parse.", Type: types.StringType, Required: true}},
		ReturnType:  types.AnyType,
		Fn:          unmarshalImpl,
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
