package arrays

import (
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func concat() types.Func {
	return types.Func{
		Name:        "concat",
		Description: "Concatenates two arrays into a single array.",
		Args: []types.ArgSpec{
			{Name: "x", Description: "The first array.", Type: types.NewArrayType(types.AnyType), Required: true},
			{Name: "y", Description: "The second array.", Type: types.NewArrayType(types.AnyType), Required: true},
		},
		ReturnType: types.NewArrayType(types.AnyType),
		Fn:         concatImpl,
	}
}

func concatImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	if x.Undefined {
		return core.U(), nil
	}

	arr1, ok := x.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf("expected %v for 'x' argument, got %v",
			types.NewArrayType(types.AnyType),
			x.Type(),
		)
	}

	y := args[1]
	if y.Undefined {
		return core.U(), nil
	}

	arr2, ok := y.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf("expected %v for 'y' argument, got %v",
			types.NewArrayType(types.AnyType),
			x.Type(),
		)
	}

	return core.V(append(arr1, arr2...)), nil
}
