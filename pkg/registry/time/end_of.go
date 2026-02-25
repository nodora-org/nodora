package time

import (
	"fmt"
	gotime "time"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func endOf() types.Func {
	return types.Func{
		Name:        "end_of",
		Description: "Returns the end of the given time unit for a timestamp (UTC).",
		Args: []types.ArgSpec{
			{Name: "ts", Description: "The timestamp in milliseconds.", Type: types.NumberType, Required: true},
			{Name: "unit", Description: "The time unit (hour, day, week, month, year).", Type: types.StringType, Required: true},
		},
		ReturnType: types.NumberType,
		Fn:         endOfImpl,
	}
}

func endOfImpl(args []core.Value) (core.Value, error) {
	tsArg := args[0]
	if tsArg.Undefined {
		return core.U(), nil
	}
	ms, ok := core.ToFloat64(tsArg.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'ts' argument, got %v", tsArg.Type())
	}

	unitArg := args[1]
	if unitArg.Undefined {
		return core.U(), nil
	}
	unit, ok := unitArg.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'unit' argument, got %v", unitArg.Type())
	}

	t := msToTime(ms)
	var result gotime.Time

	switch unit {
	case "hour":
		result = gotime.Date(t.Year(), t.Month(), t.Day(), t.Hour()+1, 0, 0, 0, gotime.UTC).Add(-gotime.Millisecond)
	case "day":
		result = gotime.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, gotime.UTC).Add(-gotime.Millisecond)
	case "week":
		weekday := int(t.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		daysUntilSunday := 7 - weekday
		d := t.AddDate(0, 0, daysUntilSunday+1)
		result = gotime.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, gotime.UTC).Add(-gotime.Millisecond)
	case "month":
		result = gotime.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, gotime.UTC).Add(-gotime.Millisecond)
	case "year":
		result = gotime.Date(t.Year()+1, gotime.January, 1, 0, 0, 0, 0, gotime.UTC).Add(-gotime.Millisecond)
	default:
		return core.U(), fmt.Errorf("unsupported unit for end_of: %q (expected hour, day, week, month, year)", unit)
	}

	return core.V(timeToMs(result)), nil
}
