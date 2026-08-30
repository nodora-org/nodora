package arrays

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func zip() types.Func {
	return types.Func{
		Name:        "zip",
		Description: "Combines two arrays into an array of pairs.",
		Args: []types.ArgSpec{
			{Name: "x", Description: "The first array.", Type: types.NewArrayType(types.AnyType), Required: true},
			{Name: "y", Description: "The second array.", Type: types.NewArrayType(types.AnyType), Required: true},
			{Name: "strict", Description: "If true, requires both arrays to have the same length.", Type: types.BoolType, Required: false},
		},
		ReturnType: types.NewArrayType(types.NewArrayType(types.AnyType)),
		Fn:         zipImpl,
		Pure:       true,
	}
}

func zipImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	if x.IsUndefined() {
		return core.U(), nil
	}

	arr1, ok := x.AsArray()
	if !ok {
		return core.U(), fmt.Errorf("expected %v for 'x' argument, got %v",
			types.NewArrayType(types.AnyType),
			x.Type(),
		)
	}

	y := args[1]
	if y.IsUndefined() {
		return core.U(), nil
	}

	arr2, ok := y.AsArray()
	if !ok {
		return core.U(), fmt.Errorf("expected %v for 'y' argument, got %v",
			types.NewArrayType(types.AnyType),
			x.Type(),
		)
	}

	strictMode := false
	if len(args) > 2 {
		strict := args[2]
		if strict.IsUndefined() {
			return core.U(), nil
		}
		strictMode, ok = strict.AsBool()
		if !ok {
			return core.U(), fmt.Errorf("expected bool for 'strict' argument, got %v", strict.Type())
		}
	}

	if strictMode && len(arr1) != len(arr2) {
		return core.U(), fmt.Errorf(
			"expected 'x' and 'y' arguments to have the same length, got %d and %d",
			len(arr1),
			len(arr2),
		)
	}

	n := min(len(arr2), len(arr1))
	result := make([]core.Value, n)
	for i := range n {
		result[i] = core.V([]core.Value{arr1[i], arr2[i]})
	}

	return core.V(result), nil
}
