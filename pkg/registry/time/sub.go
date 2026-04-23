package time

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func sub() types.Func {
	return types.Func{
		Name:        "sub",
		Description: "Subtracts a duration from a timestamp.",
		Args: []types.ArgSpec{
			{Name: "ts", Description: "The timestamp in milliseconds.", Type: types.NumberType, Required: true},
			{Name: "duration", Description: "The duration string (e.g. \"5m\", \"1h30m\", \"2d\").", Type: types.StringType, Required: true},
		},
		ReturnType: types.NumberType,
		Fn:         subImpl,
	}
}

func subImpl(args []core.Value) (core.Value, error) {
	ts := args[0]
	if ts.Undefined {
		return core.U(), nil
	}
	ms, ok := core.ToFloat64(ts.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'ts' argument, got %v", ts.Type())
	}

	dur := args[1]
	if dur.Undefined {
		return core.U(), nil
	}
	durStr, ok := dur.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'duration' argument, got %v", dur.Type())
	}

	d, err := parseDuration(durStr)
	if err != nil {
		return core.U(), err
	}

	return core.V(ms - float64(d.Milliseconds())), nil
}
