package base64

import (
	"encoding/base64"
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func isValid() types.Func {
	return types.Func{
		Name:        "is_valid",
		Description: "Returns true if the string is valid standard base64.",
		Args:        []types.ArgSpec{{Name: "str", Description: "The string to validate.", Type: types.StringType, Required: true}},
		ReturnType:  types.BoolType,
		Fn:          isValidImpl,
		Pure:        true,
	}
}

func isValidImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	_, err := base64.StdEncoding.DecodeString(strVal)
	return core.V(err == nil), nil
}
