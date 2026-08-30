package core

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func product() types.Func {
	return types.Func{
		Name:        "product",
		Description: "Returns the product of all numbers in the array.",
		Args: []types.ArgSpec{{
			Name:        "arr",
			Description: "An array of numbers to multiply.",
			Type:        types.NewArrayType(types.NumberType),
			Required:    true,
		}},
		ReturnType: types.NumberType,
		Fn:         productImpl,
		Pure:       true,
	}
}

func productImpl(args []core.Value) (core.Value, error) {
	arr := args[0]
	if arr.IsUndefined() {
		return core.U(), nil
	}

	arrVal, ok := arr.AsArray()
	if !ok {
		return core.U(), fmt.Errorf(
			"expected %v for 'arr' argument, got %v",
			types.NewArrayType(types.NumberType),
			arr.Type(),
		)
	}

	prod := float64(1)
	for _, item := range arrVal {
		if item.IsUndefined() {
			return core.U(), nil
		}
		val, ok := item.AsFloat()
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
