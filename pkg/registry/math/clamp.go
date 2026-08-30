package math

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
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
		Pure:       true,
	}
}

func clampImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	min := args[1]
	max := args[2]
	if x.IsUndefined() || min.IsUndefined() || max.IsUndefined() {
		return core.U(), nil
	}
	xx, ok := x.AsFloat()
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'x' argument, got %v", x.Type())
	}
	minV, ok := min.AsFloat()
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'min' argument, got %v", min.Type())
	}
	maxV, ok := max.AsFloat()
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
