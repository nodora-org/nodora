package json

import (
	"encoding/json"
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func unmarshal() types.Func {
	return types.Func{
		Name:       "unmarshal",
		Args:       []types.ArgSpec{{Name: "x", Types: []string{"string"}, Required: true}},
		ReturnType: "any",
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
