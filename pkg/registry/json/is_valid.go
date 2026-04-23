package json

import (
	"encoding/json"
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func isValid() types.Func {
	return types.Func{
		Name:        "is_valid",
		Description: "Returns true if the string is valid JSON.",
		Args:        []types.ArgSpec{{Name: "str", Description: "The string to validate.", Type: types.StringType, Required: true}},
		ReturnType:  types.BoolType,
		Fn:          isValidImpl,
	}
}

func isValidImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	strVal, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	var tmp any
	err := json.Unmarshal([]byte(strVal), &tmp)
	return core.V(err == nil), nil
}
