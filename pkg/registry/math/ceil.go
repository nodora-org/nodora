package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func ceil() types.Func {
	return types.Func{
		Name:        "ceil",
		Description: "Returns the smallest integer greater than or equal to x.",
		Args:        []types.ArgSpec{{Name: "x", Description: "The number to round up.", Type: types.NumberType, Required: true}},
		ReturnType:  types.NumberType,
		Fn:          ceilImpl,
	}
}

func ceilImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	if x.Undefined {
		return core.U(), nil
	}
	xx, ok := core.ToFloat64(x.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'x' argument, got %v", x.Type())
	}
	return core.V(math.Ceil(xx)), nil
}
