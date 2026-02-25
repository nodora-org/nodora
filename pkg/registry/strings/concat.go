package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func concat() types.Func {
	return types.Func{
		Name:        "concat",
		Description: "Joins an array of strings into a single string using the given delimiter.",
		Args: []types.ArgSpec{
			{
				Name:        "delim",
				Description: "The delimiter to place between each element.",
				Type:        types.StringType,
				Required:    true,
			},
			{
				Name:        "arr",
				Description: "The array of strings to join.",
				Type:        types.NewArrayType(types.StringType),
				Required:    true,
			},
		},
		ReturnType: types.StringType,
		Fn:         concatImpl,
	}
}

func concatImpl(args []core.Value) (core.Value, error) {
	delim := args[0]
	arr := args[1]
	if delim.Undefined || arr.Undefined {
		return core.U(), nil
	}
	delimVal, ok := delim.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'delim' argument, got %v", delim.Type())
	}
	arrVal, ok := arr.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf(
			"expected %v for 'arr' argument, got %v",
			types.NewArrayType(types.AnyType),
			arr.Type(),
		)
	}
	var parts []string
	for _, v := range arrVal {
		if str, ok := v.Raw.(string); ok {
			parts = append(parts, str)
		} else {
			return core.U(), fmt.Errorf("expected string in 'arr', got %v", v.Type())
		}
	}
	return core.V(strings.Join(parts, delimVal)), nil
}
