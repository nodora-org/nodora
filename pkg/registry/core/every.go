package core

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func every() types.Func {
	return types.Func{
		Name:        "every",
		Description: "Returns true if all elements in the array satisfy the predicate.",
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
		Fn:         everyImpl,
		Pure:       true,
	}
}

func everyImpl(args []core.Value) (core.Value, error) {
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

	result := true
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
		result = result && boolVal
	}

	return core.V(result), nil
}
