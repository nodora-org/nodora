package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func floor() types.Func {
	return types.Func{
		Name:       "floor",
		Args:       []types.ArgSpec{{Name: "x", Type: types.NumberType, Required: true}},
		ReturnType: types.NumberType,
		Fn:         floorImpl,
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
