package arrays

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func findIndex() types.Func {
	return types.Func{
		Name:        "find_index",
		Description: "Returns the index of the first element that satisfies the predicate, or -1 if none.",
		Args: []types.ArgSpec{
			{
				Name:        "arr",
				Description: "The array to search.",
				Type:        types.NewArrayType(types.AnyType),
				Required:    true,
			},
			{
				Name:        "fx",
				Description: "A predicate function applied to each element.",
				Type:        types.NewLambdaType([]types.Type{types.AnyType}, types.BoolType),
				Required:    true,
			},
		},
		ReturnType: types.NumberType,
		Fn:         findIndexImpl,
	}
}

func findIndexImpl(args []core.Value) (core.Value, error) {
	arr := args[0]
	if arr.Undefined {
		return core.U(), nil
	}
	arrVal, ok := arr.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf(
			"expected %v for 'arr' argument, got %v",
			types.NewArrayType(types.AnyType),
			arr.Type(),
		)
	}

	fx := args[1]
	if fx.Undefined {
		return core.U(), nil
	}

	fn, ok := fx.Raw.(*core.Lambda)
	if !ok {
		return core.U(), fmt.Errorf(
			"expected %v for 'fx' argument, got %v",
			types.NewLambdaType([]types.Type{types.AnyType}, types.BoolType),
			fx.Type(),
		)
	}

	for i, v := range arrVal {
		r, err := (*fn).Fn([]core.Value{v})
		if err != nil {
			return core.U(), err
		}
		boolVal, ok := r.Raw.(bool)
		if !ok {
			return core.U(), fmt.Errorf(
				"expected a bool value, got %v",
				r.Type(),
			)
		}
		if boolVal {
			return core.V(i), nil
		}
	}

	return core.V(-1), nil
}
