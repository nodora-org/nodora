package math

import (
	"fmt"
	"math"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func floor() types.Func {
	return types.Func{
		Name:       "floor",
		Args:       []types.ArgSpec{{Name: "x", Types: []string{"number"}, Required: true}},
		ReturnType: "number",
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
		return core.U(), fmt.Errorf("invalid x type")
	}
	return core.V(math.Floor(xx)), nil
}
