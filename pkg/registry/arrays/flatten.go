package arrays

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func flatten() types.Func {
	return types.Func{
		Name:        "flatten",
		Description: "Recursively flattens nested arrays into a single flat array.",
		Args: []types.ArgSpec{
			{Name: "arr", Description: "The array to flatten.", Type: types.NewArrayType(types.AnyType), Required: true},
		},
		ReturnType: types.NewArrayType(types.AnyType),
		Fn:         flattenImpl,
		Pure:       true,
	}
}

func flattenImpl(args []core.Value) (core.Value, error) {
	arr := args[0]
	if arr.IsUndefined() {
		return core.U(), nil
	}

	arrVal, ok := arr.AsArray()
	if !ok {
		return core.U(), fmt.Errorf("expected %v for 'arr' argument, got %v",
			types.NewArrayType(types.AnyType),
			arr.Type(),
		)
	}

	return core.V(flattenDeep(arrVal)), nil
}

func flattenDeep(values []core.Value) []core.Value {
	var result []core.Value

	for _, v := range values {
		switch rv := v.Raw.(type) {
		case []core.Value:
			result = append(result, flattenDeep(rv)...)
		default:
			result = append(result, v)
		}
	}

	return result
}
