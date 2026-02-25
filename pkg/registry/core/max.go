package core

import (
	"cmp"
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func max() types.Func {
	return types.Func{
		Name: "max",
		Args: []types.ArgSpec{{
			Name: "value",
			Type: types.NewUnionType(
				types.NewArrayType(types.NumberType),
				types.NewArrayType(types.StringType),
			),
			Required: true,
		}},
		ReturnType: types.NewUnionType(types.NumberType, types.StringType),
		Fn:         maxImpl,
	}
}

func maxImpl(args []core.Value) (core.Value, error) {
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
		max, ok := findMax[string](arr)
		if !ok {
			return core.U(), nil
		}
		return core.V(max), nil
	case float64:
		max, ok := findMax[float64](arr)
		if !ok {
			return core.U(), nil
		}
		return core.V(max), nil
	default:
		return core.U(), nil
	}
}

func findMax[T cmp.Ordered](arr []core.Value) (T, bool) {
	var max T
	for i, item := range arr {
		val, ok := item.Raw.(T)
		if !ok || item.Undefined {
			return max, false
		}
		if i == 0 || val > max {
			max = val
		}
	}
	return max, true
}
