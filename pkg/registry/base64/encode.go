package base64

import (
	"encoding/base64"
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func encode() types.Func {
	return types.Func{
		Name:        "encode",
		Description: "Encodes a string to standard base64.",
		Args:        []types.ArgSpec{{Name: "str", Description: "The string to encode.", Type: types.StringType, Required: true}},
		ReturnType:  types.StringType,
		Fn:         encodeImpl,
	}
}

func encodeImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(strVal))
	return core.V(encoded), nil
}
