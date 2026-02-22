package base64url

import (
	"encoding/base64"
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func decode() types.Func {
	return types.Func{
		Name:       "decode",
		Args:       []types.ArgSpec{{Name: "str", Type: types.StringType, Required: true}},
		ReturnType: types.StringType,
		Fn:         decodeImpl,
	}
}

func decodeImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	decoded, err := base64.URLEncoding.DecodeString(strVal)
	if err != nil {
		return core.U(), err
	}
	return core.V(string(decoded)), nil
}
