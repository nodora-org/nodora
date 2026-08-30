package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func round() types.Func {
	return types.Func{
		Name:        "round",
		Description: "Rounds a number to the nearest integer, or to the given number of decimal places.",
		Args: []types.ArgSpec{
			{Name: "x", Description: "The number to round.", Type: types.NumberType, Required: true},
			{Name: "decimals", Description: "The number of decimal places to round to. Defaults to 0.", Type: types.NumberType, Required: false},
		},
		ReturnType: types.NumberType,
		Fn:         roundImpl,
		Pure:       true,
	}
}

func roundImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	if x.IsUndefined() {
		return core.U(), nil
	}
	xx, ok := x.AsFloat()
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'x' argument, got %v", x.Type())
	}

	decimals := 0
	if len(args) > 1 {
		d := args[1]
		if d.IsUndefined() {
			return core.U(), nil
		}
		decimals, ok = d.AsInt()
		if !ok {
			return core.U(), fmt.Errorf("expected number for 'decimals' argument, got %v", d.Type())
		}
	}

	if decimals == 0 {
		return core.V(math.Round(xx)), nil
	}

	factor := math.Pow(10, float64(decimals))
	return core.V(math.Round(xx*factor) / factor), nil
}
