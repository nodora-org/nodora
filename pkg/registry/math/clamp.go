package math

import (
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func clamp() types.Func {
	return types.Func{
		Name:        "clamp",
		Description: "Clamps a number to the range [min, max].",
		Args: []types.ArgSpec{
			{Name: "x", Description: "The number to clamp.", Type: types.NumberType, Required: true},
			{Name: "min", Description: "The lower bound.", Type: types.NumberType, Required: true},
			{Name: "max", Description: "The upper bound.", Type: types.NumberType, Required: true},
		},
		ReturnType: types.NumberType,
		Fn:         clampImpl,
	}
}

func clampImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	min := args[1]
	max := args[2]
	if x.Undefined || min.Undefined || max.Undefined {
		return core.U(), nil
	}
	xx, ok := core.ToFloat64(x.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'x' argument, got %v", x.Type())
	}
	minV, ok := core.ToFloat64(min.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'min' argument, got %v", min.Type())
	}
	maxV, ok := core.ToFloat64(max.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'max' argument, got %v", max.Type())
	}
	if xx < minV {
		return core.V(minV), nil
	}
	if xx > maxV {
		return core.V(maxV), nil
	}
	return core.V(xx), nil
}
