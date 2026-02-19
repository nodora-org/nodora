package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func ceil() types.Func {
	return types.Func{
		Name:       "ceil",
		Args:       []types.ArgSpec{{Name: "x", Type: types.NumberType, Required: true}},
		ReturnType: types.NumberType,
		Fn:         ceilImpl,
	}
}

func ceilImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	if x.Undefined {
		return core.U(), nil
	}
	xx, ok := core.ToFloat64(x.Raw)
	if !ok {
		return core.U(), fmt.Errorf("invalid x type")
	}
	return core.V(math.Ceil(xx)), nil
}
