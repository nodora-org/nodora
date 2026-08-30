package base64url

import (
	"encoding/base64"
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func encode() types.Func {
	return types.Func{
		Name:        "encode",
		Description: "Encodes a string to URL-safe base64.",
		Args:        []types.ArgSpec{{Name: "str", Description: "The string to encode.", Type: types.StringType, Required: true}},
		ReturnType:  types.StringType,
		Fn:          encodeImpl,
		Pure:        true,
	}
}

func encodeImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	strVal, ok := str.AsString()
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	encoded := base64.URLEncoding.EncodeToString([]byte(strVal))
	return core.V(encoded), nil
}
