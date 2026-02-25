package core

import (
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func product() types.Func {
	return types.Func{
		Name: "product",
		Args: []types.ArgSpec{{
			Name:     "arr",
			Type:     types.NewArrayType(types.NumberType),
			Required: true,
		}},
		ReturnType: types.NumberType,
		Fn:         productImpl,
	}
}

func productImpl(args []core.Value) (core.Value, error) {
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

	prod := float64(1)
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
		prod *= val
	}

	return core.V(prod), nil
}
