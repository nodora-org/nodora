package core

import (
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func sum() types.Func {
	return types.Func{
		Name: "sum",
		Args: []types.ArgSpec{{
			Name:     "arr",
			Type:     types.NewArrayType(types.NumberType),
			Required: true,
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
