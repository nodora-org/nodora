package base64

import (
	"encoding/base64"
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func decode() types.Func {
	return types.Func{
		Name:        "decode",
		Description: "Decodes a standard base64 encoded string.",
		Args:        []types.ArgSpec{{Name: "str", Description: "The base64 encoded string to decode.", Type: types.StringType, Required: true}},
		ReturnType:  types.StringType,
		Fn:          decodeImpl,
		Pure:        true,
	}
}

func decodeImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	decoded, err := base64.StdEncoding.DecodeString(strVal)
	if err != nil {
		return core.U(), err
	}
	return core.V(string(decoded)), nil
}
