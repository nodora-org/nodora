package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func maxFunc() types.Func {
	return types.Func{
		Name: "max",
		Args: []types.ArgSpec{
			{Name: "x", Type: types.NumberType, Required: true},
			{Name: "y", Type: types.NumberType, Required: true},
		},
		ReturnType: types.NumberType,
		Fn:         maxImpl,
	}
}

func maxImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	y := args[1]
	if x.Undefined || y.Undefined {
		return core.U(), nil
	}
	xx, ok := core.ToFloat64(x.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'x' argument, got %v", x.Type())
	}
	yy, ok := core.ToFloat64(y.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'y' argument, got %v", y.Type())
	}
	return core.V(math.Max(xx, yy)), nil
}
