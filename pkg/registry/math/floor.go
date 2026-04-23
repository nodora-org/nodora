package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func floor() types.Func {
	return types.Func{
		Name:        "floor",
		Description: "Returns the largest integer less than or equal to x.",
		Args:        []types.ArgSpec{{Name: "x", Description: "The number to round down.", Type: types.NumberType, Required: true}},
		ReturnType:  types.NumberType,
		Fn:          floorImpl,
	}
}

func floorImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	if x.Undefined {
		return core.U(), nil
	}
	xx, ok := core.ToFloat64(x.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'x' argument, got %v", x.Type())
	}
	return core.V(math.Floor(xx)), nil
}
