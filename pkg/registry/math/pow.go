package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func pow() types.Func {
	return types.Func{
		Name:        "pow",
		Description: "Returns x raised to the power of y.",
		Args: []types.ArgSpec{
			{Name: "x", Description: "The base.", Type: types.NumberType, Required: true},
			{Name: "y", Description: "The exponent.", Type: types.NumberType, Required: true},
		},
		ReturnType: types.NumberType,
		Fn:         powImpl,
		Pure:       true,
	}
}

func powImpl(args []core.Value) (core.Value, error) {
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
	return core.V(math.Pow(xx, yy)), nil
}
