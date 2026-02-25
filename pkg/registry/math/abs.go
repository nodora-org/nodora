package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func abs() types.Func {
	return types.Func{
		Name:        "abs",
		Description: "Returns the absolute value of a number.",
		Args:        []types.ArgSpec{{Name: "x", Description: "The number.", Type: types.NumberType, Required: true}},
		ReturnType:  types.NumberType,
		Fn:         absImpl,
	}
}

func absImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	if x.Undefined {
		return core.U(), nil
	}
	xx, ok := core.ToFloat64(x.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'x' argument, got %v", x.Type())
	}
	return core.V(math.Abs(xx)), nil
}
