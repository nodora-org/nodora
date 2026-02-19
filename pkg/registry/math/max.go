package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func max() types.Func {
	return types.Func{
		Name: "max",
		Args: []types.ArgSpec{
			{Name: "x", Types: []string{"number"}, Required: true},
			{Name: "y", Types: []string{"number"}, Required: true},
		},
		ReturnType: "number",
		Fn:         maxImpl,
	}
}

func maxImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	y := args[1]
	if x.Undefined || y.Undefined {
		return core.U(), nil
	}
	xx, ok := core.ToFloat64(x.Raw)
	if !ok {
		return core.U(), fmt.Errorf("invalid x type")
	}
	yy, ok := core.ToFloat64(y.Raw)
	if !ok {
		return core.U(), fmt.Errorf("invalid y type")
	}
	return core.V(math.Max(xx, yy)), nil
}
