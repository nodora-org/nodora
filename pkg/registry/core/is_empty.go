package core

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func isEmpty() types.Func {
	return types.Func{
		Name:        "is_empty",
		Description: "Returns true if the array has no elements.",
		Args: []types.ArgSpec{{
			Name:        "arr",
			Description: "The array to check.",
			Type:        types.NewArrayType(types.AnyType),
			Required:    true,
		}},
		ReturnType: types.BoolType,
		Fn:         isEmptyImpl,
		Pure:       true,
	}
}

func isEmptyImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.IsUndefined() {
		return core.U(), nil
	}
	arr, ok := value.AsArray()
	if !ok {
		return core.U(), fmt.Errorf(
			"expected %s for 'arr' argument, got %v",
			types.NewArrayType(types.AnyType),
			value.Type(),
		)
	}
	return core.V(len(arr) == 0), nil
}
