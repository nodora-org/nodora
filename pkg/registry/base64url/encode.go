package base64url

import (
	"encoding/base64"
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func encode() types.Func {
	return types.Func{
		Name:       "encode",
		Args:       []types.ArgSpec{{Name: "str", Types: []string{"string"}, Required: true}},
		ReturnType: "string",
		Fn:         encodeImpl,
	}
}

func encodeImpl(args []core.Value) (core.Value, error) {
	str, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid str type")
	}
	encoded := base64.URLEncoding.EncodeToString([]byte(str))
	return core.V(encoded), nil
}
