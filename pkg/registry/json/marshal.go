package json

import (
	"encoding/json"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func marshal() types.Func {
	return types.Func{
		Name:       "marshal",
		Args:       []types.ArgSpec{{Name: "x", Types: nil, Required: true}},
		ReturnType: "string",
		Fn:         marshalImpl,
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
