package arrays

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func mapFunc() types.Func {
	t := types.NewTypeVar("T")
	u := types.NewTypeVar("U")
	return types.Func{
		Name:        "map",
		Description: "Returns a new array with the function applied to each element.",
		Args: []types.ArgSpec{
			{
				Name:        "arr",
				Description: "The array to transform.",
				Type:        types.NewArrayType(t),
				Required:    true,
			},
			{
				Name:        "fx",
				Description: "A function applied to each element to produce the new value.",
				Type:        types.NewLambdaType([]types.Type{t}, u),
				Required:    true,
			},
		},
		ReturnType: types.NewArrayType(u),
		Fn:         mapImpl,
		Pure:       true,
	}
}

func mapImpl(args []core.Value) (core.Value, error) {
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
			types.NewLambdaType([]types.Type{types.AnyType}, types.AnyType),
			fx.Type(),
		)
	}

	result := make([]core.Value, len(arrVal))
	for i, v := range arrVal {
		r, err := (*fn).Fn([]core.Value{v})
		if err != nil {
			return core.U(), err
		}
		if r.Undefined {
			return core.U(), nil
		}
		result[i] = r
	}

	return core.V(result), nil
}
