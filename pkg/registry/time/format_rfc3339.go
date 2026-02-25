package time

import (
	"fmt"
	gotime "time"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func formatRFC3339() types.Func {
	return types.Func{
		Name:        "format_rfc3339",
		Description: "Formats a timestamp in milliseconds as an RFC3339 date string.",
		Args: []types.ArgSpec{
			{Name: "ts", Description: "The timestamp in milliseconds.", Type: types.NumberType, Required: true},
		},
		ReturnType: types.StringType,
		Fn:         formatRFC3339Impl,
	}
}

func formatRFC3339Impl(args []core.Value) (core.Value, error) {
	v := args[0]
	if v.Undefined {
		return core.U(), nil
	}
	ms, ok := core.ToFloat64(v.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'ts' argument, got %v", v.Type())
	}

	t := gotime.UnixMilli(int64(ms)).UTC()
	return core.V(t.Format(gotime.RFC3339Nano)), nil
}
