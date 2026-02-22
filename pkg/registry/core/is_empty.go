package core

import (
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func isEmpty() types.Func {
	return types.Func{
		Name: "is_empty",
		Args: []types.ArgSpec{{
			Name:     "arr",
			Type:     types.NewArrayType(types.AnyType),
			Required: true,
		}},
		ReturnType: types.BoolType,
		Fn:         isEmptyImpl,
	}
}

func isEmptyImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	arr, ok := value.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf(
			"expected %s for 'arr' argument, got %v",
			types.NewArrayType(types.AnyType),
			value.Type(),
		)
	}
	return core.V(len(arr) == 0), nil
}
