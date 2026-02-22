package arrays

import (
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func slice() types.Func {
	return types.Func{
		Name: "slice",
		Args: []types.ArgSpec{
			{Name: "arr", Type: types.NewArrayType(types.AnyType), Required: true},
			{Name: "start", Type: types.NumberType, Required: true},
			{Name: "stop", Type: types.NumberType, Required: false},
		},
		ReturnType: types.NewArrayType(types.AnyType),
		Fn:         sliceImpl,
	}
}

func sliceImpl(args []core.Value) (core.Value, error) {
	arr := args[0]
	if arr.Undefined {
		return core.U(), nil
	}

	arrVal, ok := arr.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf("expected %v for 'arr' argument, got %v",
			types.NewArrayType(types.AnyType),
			arr.Type(),
		)
	}

	start := args[1]
	startVal, ok := core.ToInt(start.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'start' argument, got %v",
			start.Type(),
		)
	}

	startVal = clamp(startVal, 0, len(arrVal))
	stopVal := len(arrVal)

	if len(args) > 2 {
		stop := args[2]
		stopVal, ok = core.ToInt(stop.Raw)
		if !ok {
			return core.U(), fmt.Errorf("expected number for 'stop' argument, got %v",
				stop.Type(),
			)
		}
		stopVal = clamp(stopVal, 0, len(arrVal))
	}

	if startVal >= stopVal {
		return core.V([]core.Value{}), nil
	}

	return core.V(arrVal[startVal:stopVal]), nil
}

func clamp(x, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}
