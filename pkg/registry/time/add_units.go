package time

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func addUnits() types.Func {
	return types.Func{
		Name:        "add_units",
		Description: "Adds a number of time units to a timestamp. Unit defaults to \"ms\" if not provided.",
		Args: []types.ArgSpec{
			{Name: "ts", Description: "The timestamp in milliseconds.", Type: types.NumberType, Required: true},
			{Name: "n", Description: "The number of units to add.", Type: types.NumberType, Required: true},
			{Name: "unit", Description: "The unit suffix (ms, s, m, h, d, w). Defaults to \"ms\".", Type: types.StringType, Required: false},
		},
		ReturnType: types.NumberType,
		Fn:         addUnitsImpl,
	}
}

func addUnitsImpl(args []core.Value) (core.Value, error) {
	ts := args[0]
	if ts.Undefined {
		return core.U(), nil
	}
	ms, ok := core.ToFloat64(ts.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'ts' argument, got %v", ts.Type())
	}

	nArg := args[1]
	if nArg.Undefined {
		return core.U(), nil
	}
	n, ok := core.ToFloat64(nArg.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'n' argument, got %v", nArg.Type())
	}

	unit := "ms"
	if len(args) > 2 && !args[2].Undefined {
		u, ok := args[2].Raw.(string)
		if !ok {
			return core.U(), fmt.Errorf("expected string for 'unit' argument, got %v", args[2].Type())
		}
		unit = u
	}

	d, err := unitToDuration(unit)
	if err != nil {
		return core.U(), err
	}

	return core.V(ms + n*float64(d.Milliseconds())), nil
}
