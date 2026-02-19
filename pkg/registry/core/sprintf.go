package core

import (
	"fmt"
	"math"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func sprintf() types.Func {
	return types.Func{
		Name: "sprintf",
		Args: []types.ArgSpec{
			{Name: "fmt", Type: types.StringType, Required: true},
			{Name: "args", Type: types.NewArrayType(types.AnyType), Required: true},
		},
		ReturnType: types.StringType,
		Fn:         sprintfImpl,
	}
}

func sprintfImpl(args []core.Value) (core.Value, error) {
	fmtStr, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid fmt type")
	}
	arr, ok := args[1].Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf("invalid args type")
	}

	interfaceArgs := make([]any, len(arr))
	for i, v := range arr {
		// format compatibility
		if f, ok := v.Raw.(float64); ok && f == math.Trunc(f) && !math.IsInf(f, 0) {
			interfaceArgs[i] = int64(f)
		} else {
			interfaceArgs[i] = v.ToRaw()
		}
	}

	result := fmt.Sprintf(fmtStr, interfaceArgs...)
	return core.V(result), nil
}
