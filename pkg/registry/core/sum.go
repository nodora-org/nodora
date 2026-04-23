package core

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func sum() types.Func {
	return types.Func{
		Name:        "sum",
		Description: "Returns the sum of all numbers in the array.",
		Args: []types.ArgSpec{{
			Name:        "arr",
			Description: "An array of numbers to sum.",
			Type:        types.NewArrayType(types.NumberType),
			Required:    true,
		}},
		ReturnType: types.NumberType,
		Fn:         sumImpl,
	}
}

func sumImpl(args []core.Value) (core.Value, error) {
	arr := args[0]
	if arr.Undefined {
		return core.U(), nil
	}

	arrVal, ok := arr.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf(
			"expected %v for 'arr' argument, got %v",
			types.NewArrayType(types.NumberType),
			arr.Type(),
		)
	}

	sum := float64(0)
	for _, item := range arrVal {
		if item.Undefined {
			return core.U(), nil
		}
		val, ok := core.ToFloat64(item.Raw)
		if !ok {
			return core.U(), fmt.Errorf(
				"expected element of type number, got %v",
				item.Type(),
			)
		}
		sum += val
	}

	return core.V(sum), nil
}
