package core

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func some() types.Func {
	return types.Func{
		Name:        "some",
		Description: "Returns true if at least one element in the array satisfies the predicate.",
		Args: []types.ArgSpec{
			{
				Name:        "arr",
				Description: "The array to test.",
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
		ReturnType: types.BoolType,
		Fn:         someImpl,
		Pure:       true,
	}
}

func someImpl(args []core.Value) (core.Value, error) {
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

	result := false
	for _, v := range arrVal {
		r, err := (*fn).Fn([]core.Value{v})
		if err != nil {
			return core.U(), err
		}
		if r.Undefined {
			return core.U(), nil
		}
		boolVal, ok := r.Raw.(bool)
		if !ok {
			return core.U(), fmt.Errorf(
				"expected a bool value, got %v",
				r.Type(),
			)
		}
		result = result || boolVal
	}

	return core.V(result), nil
}
