package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func maxFunc() types.Func {
	return types.Func{
		Name:        "max",
		Description: "Returns the larger of two numbers.",
		Args: []types.ArgSpec{
			{Name: "x", Description: "The first number.", Type: types.NumberType, Required: true},
			{Name: "y", Description: "The second number.", Type: types.NumberType, Required: true},
		},
		ReturnType: types.NumberType,
		Fn:         maxImpl,
		Pure:       true,
	}
}

func maxImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	y := args[1]
	if x.IsUndefined() || y.IsUndefined() {
		return core.U(), nil
	}
	xx, ok := x.AsFloat()
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'x' argument, got %v", x.Type())
	}
	yy, ok := y.AsFloat()
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'y' argument, got %v", y.Type())
	}
	return core.V(math.Max(xx, yy)), nil
}
