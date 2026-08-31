package time

import (
	"fmt"
	gotime "time"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func startOf() types.Func {
	return types.Func{
		Name:        "start_of",
		Description: "Returns the start of the given time unit for a timestamp (UTC).",
		Args: []types.ArgSpec{
			{Name: "ts", Description: "The timestamp in milliseconds.", Type: types.NumberType, Required: true},
			{Name: "unit", Description: "The time unit (hour, day, week, month, year).", Type: types.StringType, Required: true},
		},
		ReturnType: types.NumberType,
		Fn:         startOfImpl,
		Pure:       true,
	}
}

func startOfImpl(args []core.Value) (core.Value, error) {
	tsArg := args[0]
	if tsArg.IsUndefined() {
		return core.U(), nil
	}
	ms, ok := tsArg.AsFloat()
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'ts' argument, got %v", tsArg.Type())
	}

	unitArg := args[1]
	if unitArg.IsUndefined() {
		return core.U(), nil
	}
	unit, ok := unitArg.AsString()
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'unit' argument, got %v", unitArg.Type())
	}

	t := msToTime(ms)
	var result gotime.Time

	switch unit {
	case "hour":
		result = gotime.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, gotime.UTC)
	case "day":
		result = gotime.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, gotime.UTC)
	case "week":
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		d := t.AddDate(0, 0, -(weekday - 1))
		result = gotime.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, gotime.UTC)
	case "month":
		result = gotime.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, gotime.UTC)
	case "year":
		result = gotime.Date(t.Year(), gotime.January, 1, 0, 0, 0, 0, gotime.UTC)
	default:
		return core.U(), fmt.Errorf("unsupported unit for start_of: %q (expected hour, day, week, month, year)", unit)
	}

	return core.V(timeToMs(result)), nil
}
