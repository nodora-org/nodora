package core

import (
	"fmt"
	"math"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func sprintf() types.Func {
	return types.Func{
		Name:        "sprintf",
		Description: "Formats a string using a format specifier and arguments.",
		Args: []types.ArgSpec{
			{Name: "fmt", Description: "The format string.", Type: types.StringType, Required: true},
			{Name: "args", Description: "The values to substitute into the format string.", Type: types.NewArrayType(types.AnyType), Required: false},
		},
		ReturnType: types.StringType,
		Fn:         sprintfImpl,
		Pure:       true,
	}
}

func sprintfImpl(args []core.Value) (core.Value, error) {
	fmt_ := args[0]
	fmtVal, ok := fmt_.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf(
			"expected string for 'fmt' argument, got %v",
			fmt_.Type(),
		)
	}

	arrVal := make([]core.Value, 0)
	if len(args) > 1 {
		args := args[1]
		arrVal, ok = args.Raw.([]core.Value)
		if !ok {
			return core.U(), fmt.Errorf(
				"expected %v for 'args' argument, got %v",
				types.NewArrayType(types.AnyType),
				args.Type(),
			)
		}
	}

	interfaceArgs := make([]any, len(arrVal))
	for i, v := range arrVal {
		// format compatibility
		if f, ok := v.Raw.(float64); ok && f == math.Trunc(f) && !math.IsInf(f, 0) {
			interfaceArgs[i] = int64(f)
		} else {
			interfaceArgs[i] = v.ToRaw()
		}
	}

	result := fmt.Sprintf(fmtVal, interfaceArgs...)
	return core.V(result), nil
}
