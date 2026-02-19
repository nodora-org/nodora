package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func concat() types.Func {
	return types.Func{
		Name: "concat",
		Args: []types.ArgSpec{
			{
				Name:     "delim",
				Types:    []string{"string"},
				Required: true,
			},
			{
				Name:     "arr",
				Types:    []string{"array"},
				Required: true,
			},
		},
		ReturnType: "string",
		Fn:         concatImpl,
	}
}

func concatImpl(args []core.Value) (core.Value, error) {
	delimValue := args[0]
	arrValue := args[1]
	if delimValue.Undefined || arrValue.Undefined {
		return core.U(), nil
	}
	delim, ok := delimValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for delim", delimValue.Type())
	}
	arr, ok := arrValue.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for arr", arrValue.Type())
	}
	var parts []string
	for _, v := range arr {
		if str, ok := v.Raw.(string); ok {
			parts = append(parts, str)
		} else {
			return core.U(), fmt.Errorf("array element type %v not supported", v.Type())
		}
	}
	return core.V(strings.Join(parts, delim)), nil
}
