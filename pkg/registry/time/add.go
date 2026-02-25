package time

import (
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func add() types.Func {
	return types.Func{
		Name:        "add",
		Description: "Adds a duration to a timestamp. Duration uses shorthand suffixes (e.g. \"5m\", \"1h30m\", \"2d\").",
		Args: []types.ArgSpec{
			{Name: "ts", Description: "The timestamp in milliseconds.", Type: types.NumberType, Required: true},
			{Name: "duration", Description: "The duration string (e.g. \"5m\", \"1h30m\", \"2d\").", Type: types.StringType, Required: true},
		},
		ReturnType: types.NumberType,
		Fn:         addImpl,
	}
}

func addImpl(args []core.Value) (core.Value, error) {
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

	return core.V(ms + float64(d.Milliseconds())), nil
}
