package json

import (
	"encoding/json"
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func isValid() types.Func {
	return types.Func{
		Name:       "is_valid",
		Args:       []types.ArgSpec{{Name: "x", Types: []string{"string"}, Required: true}},
		ReturnType: "bool",
		Fn:         isValidImpl,
	}
}

func isValidImpl(args []core.Value) (core.Value, error) {
	str, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid x type")
	}
	var tmp any
	err := json.Unmarshal([]byte(str), &tmp)
	return core.V(err == nil), nil
}
