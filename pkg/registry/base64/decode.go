package base64

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
	str, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid str type")
	}
	decoded, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return core.U(), err
	}
	return core.V(string(decoded)), nil
}
