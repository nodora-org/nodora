package math

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func clamp() types.Func {
	return types.Func{
		Name: "clamp",
		Args: []types.ArgSpec{
			{Name: "x", Types: []string{"number"}, Required: true},
			{Name: "min", Types: []string{"number"}, Required: true},
			{Name: "max", Types: []string{"number"}, Required: true},
		},
		ReturnType: "number",
		Fn:         clampImpl,
	}
}

func clampImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	minVal := args[1]
	maxVal := args[2]
	if x.Undefined || minVal.Undefined || maxVal.Undefined {
		return core.U(), nil
	}
	xx, ok := core.ToFloat64(x.Raw)
	if !ok {
		return core.U(), fmt.Errorf("invalid x type")
	}
	minV, ok := core.ToFloat64(minVal.Raw)
	if !ok {
		return core.U(), fmt.Errorf("invalid min type")
	}
	maxV, ok := core.ToFloat64(maxVal.Raw)
	if !ok {
		return core.U(), fmt.Errorf("invalid max type")
	}
	if xx < minV {
		return core.V(minV), nil
	}
	if xx > maxV {
		return core.V(maxV), nil
	}
	return core.V(xx), nil
}
