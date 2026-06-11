package arrays

import (
	"fmt"
	"slices"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func reverse() types.Func {
	return types.Func{
		Name:        "reverse",
		Description: "Reverses the order of elements in an array.",
		Args: []types.ArgSpec{
			{Name: "arr", Description: "The array to reverse.", Type: types.NewArrayType(types.AnyType), Required: true},
		},
		ReturnType: types.NewArrayType(types.AnyType),
		Fn:         reverseImpl,
		Pure:       true,
	}
}

func reverseImpl(args []core.Value) (core.Value, error) {
	arr := args[0]
	if arr.Undefined {
		return core.U(), nil
	}

	arrVal, ok := arr.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf("expected %v for 'arr' argument, got %v",
			types.NewArrayType(types.AnyType),
			arr.Type(),
		)
	}

	slices.Reverse(arrVal)
	return core.V(arrVal), nil
}
