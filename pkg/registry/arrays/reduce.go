package arrays

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func reduce() types.Func {
	t := types.NewTypeVar("T")
	u := types.NewTypeVar("U")
	return types.Func{
		Name:        "reduce",
		Description: "Reduces the array to a single value by applying an accumulator function left to right.",
		Args: []types.ArgSpec{
			{
				Name:        "arr",
				Description: "The array to reduce.",
				Type:        types.NewArrayType(t),
				Required:    true,
			},
			{
				Name:        "fx",
				Description: "An accumulator function called with (acc, element) returning the next accumulator.",
				Type:        types.NewLambdaType([]types.Type{u, t}, u),
				Required:    true,
			},
			{
				Name:        "init",
				Description: "The initial accumulator value.",
				Type:        u,
				Required:    true,
			},
		},
		ReturnType: u,
		Fn:         reduceImpl,
		Pure:       true,
	}
}

func reduceImpl(args []core.Value) (core.Value, error) {
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
			types.NewLambdaType([]types.Type{types.AnyType, types.AnyType}, types.AnyType),
			fx.Type(),
		)
	}

	acc := args[2]
	for _, v := range arrVal {
		r, err := (*fn).Fn([]core.Value{acc, v})
		if err != nil {
			return core.U(), err
		}
		acc = r
	}

	return acc, nil
}
