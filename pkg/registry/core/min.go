package core

import (
	"cmp"
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func min() types.Func {
	return types.Func{
		Name:        "min",
		Description: "Returns the minimum element from an array of numbers or strings.",
		Args: []types.ArgSpec{{
			Name:        "value",
			Description: "An array of numbers or strings to find the minimum of.",
			Type: types.NewUnionType(
				types.NewArrayType(types.NumberType),
				types.NewArrayType(types.StringType),
			),
			Required: true,
		}},
		ReturnType: types.NewUnionType(types.NumberType, types.StringType),
		Fn:         minImpl,
	}
}

func minImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}

	arr, ok := value.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf(
			"expected %v for 'value' argument, got %v",
			types.NewUnionType(types.NewArrayType(types.NumberType), types.NewArrayType(types.StringType)),
			value.Type(),
		)
	}

	if len(arr) == 0 {
		return core.U(), nil
	}

	switch arr[0].Raw.(type) {
	case string:
		min, ok := findMin[string](arr)
		if !ok {
			return core.U(), nil
		}
		return core.V(min), nil
	case float64:
		min, ok := findMin[float64](arr)
		if !ok {
			return core.U(), nil
		}
		return core.V(min), nil
	default:
		return core.U(), nil
	}
}

func findMin[T cmp.Ordered](arr []core.Value) (T, bool) {
	var min T
	for i, item := range arr {
		val, ok := item.Raw.(T)
		if !ok || item.Undefined {
			return min, false
		}
		if i == 0 || val < min {
			min = val
		}
	}
	return min, true
}
