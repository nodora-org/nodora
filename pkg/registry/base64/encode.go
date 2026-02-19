package base64

import (
	"encoding/base64"
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func encode() types.Func {
	return types.Func{
		Name:       "encode",
		Args:       []types.ArgSpec{{Name: "str", Type: types.StringType, Required: true}},
		ReturnType: types.StringType,
		Fn:         encodeImpl,
	}
}

func encodeImpl(args []core.Value) (core.Value, error) {
	str, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid str type")
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(str))
	return core.V(encoded), nil
}
