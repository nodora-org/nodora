package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func split() types.Func {
	return types.Func{
		Name: "split",
		Args: []types.ArgSpec{
			{
				Name:     "str",
				Type:     types.StringType,
				Required: true,
			},
			{
				Name:     "delim",
				Type:     types.StringType,
				Required: true,
			},
		},
		ReturnType: types.NewArrayType(types.StringType),
		Fn:         splitImpl,
	}
}

func splitImpl(args []core.Value) (core.Value, error) {
	strValue := args[0]
	delimValue := args[1]
	if strValue.Undefined || delimValue.Undefined {
		return core.U(), nil
	}
	str, ok := strValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for str", strValue.Type())
	}
	delim, ok := delimValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for delim", delimValue.Type())
	}
	parts := strings.Split(str, delim)
	result := make([]core.Value, len(parts))
	for i, part := range parts {
		result[i] = core.V(part)
	}
	return core.V(result), nil
}
