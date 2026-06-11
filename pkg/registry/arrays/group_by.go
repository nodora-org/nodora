package arrays

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func groupBy() types.Func {
	return types.Func{
		Name:        "group_by",
		Description: "Groups array elements into an object by the result of a key function.",
		Args: []types.ArgSpec{
			{
				Name:        "arr",
				Description: "The array to group.",
				Type:        types.NewArrayType(types.AnyType),
				Required:    true,
			},
			{
				Name:        "fx",
				Description: "A function that returns the grouping key for each element.",
				Type:        types.NewLambdaType([]types.Type{types.AnyType}, types.AnyType),
				Required:    true,
			},
		},
		ReturnType: types.ObjectType,
		Fn:         groupByImpl,
		Pure:       true,
	}
}

func groupByImpl(args []core.Value) (core.Value, error) {
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

	result := make(core.ValueMap)
	for _, v := range arrVal {
		r, err := (*fn).Fn([]core.Value{v})
		if err != nil {
			return core.U(), err
		}
		key := fmt.Sprintf("%v", r.ToRaw())
		if group, ok := result[key]; ok {
			result[key] = core.V(append(group.Raw.([]core.Value), v))
		} else {
			result[key] = core.V([]core.Value{v})
		}
	}

	return core.V(result), nil
}
