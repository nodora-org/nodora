package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func sqrt() types.Func {
	return types.Func{
		Name:        "sqrt",
		Description: "Returns the square root of x. Yields undefined for negative inputs.",
		Args:        []types.ArgSpec{{Name: "x", Description: "The number.", Type: types.NumberType, Required: true}},
		ReturnType:  types.NumberType,
		Fn:          sqrtImpl,
		Pure:        true,
	}
}

func sqrtImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	if x.Undefined {
		return core.U(), nil
	}
	xx, ok := core.ToFloat64(x.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'x' argument, got %v", x.Type())
	}
	if xx < 0 {
		return core.U(), nil
	}
	return core.V(math.Sqrt(xx)), nil
}
