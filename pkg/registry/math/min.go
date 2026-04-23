package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func minFunc() types.Func {
	return types.Func{
		Name:        "min",
		Description: "Returns the smaller of two numbers.",
		Args: []types.ArgSpec{
			{Name: "x", Description: "The first number.", Type: types.NumberType, Required: true},
			{Name: "y", Description: "The second number.", Type: types.NumberType, Required: true},
		},
		ReturnType: types.NumberType,
		Fn:         minImpl,
	}
}

func minImpl(args []core.Value) (core.Value, error) {
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
	return core.V(math.Min(xx, yy)), nil
}
