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
		Args:       []types.ArgSpec{{Name: "x", Type: types.StringType, Required: true}},
		ReturnType: types.AnyType,
		Fn:         unmarshalImpl,
	}
}

func unmarshalImpl(args []core.Value) (core.Value, error) {
	str, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid x type")
	}
	var raw any
	err := json.Unmarshal([]byte(str), &raw)
	if err != nil {
		return core.U(), err
	}
	return core.NormalizeValue(raw), nil
}
