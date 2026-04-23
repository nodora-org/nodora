package json

import (
	"encoding/json"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func marshal() types.Func {
	return types.Func{
		Name:        "marshal",
		Description: "Serializes a value to a JSON string.",
		Args:        []types.ArgSpec{{Name: "x", Description: "The value to serialize.", Type: types.AnyType, Required: true}},
		ReturnType:  types.StringType,
		Fn:          marshalImpl,
	}
}

func marshalImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	data, err := json.Marshal(value.ToRaw())
	if err != nil {
		return core.U(), err
	}
	return core.V(string(data)), nil
}
