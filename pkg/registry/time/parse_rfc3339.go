package time

import (
	"fmt"
	gotime "time"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func parseRFC3339() types.Func {
	return types.Func{
		Name:        "parse_rfc3339",
		Description: "Parses an RFC3339 date string and returns the timestamp in milliseconds.",
		Args: []types.ArgSpec{
			{Name: "value", Description: "The RFC3339 date string.", Type: types.StringType, Required: true},
		},
		ReturnType: types.NumberType,
		Fn:         parseRFC3339Impl,
	}
}

func parseRFC3339Impl(args []core.Value) (core.Value, error) {
	v := args[0]
	if v.Undefined {
		return core.U(), nil
	}
	s, ok := v.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'value' argument, got %v", v.Type())
	}

	t, err := gotime.Parse(gotime.RFC3339Nano, s)
	if err != nil {
		return core.U(), fmt.Errorf("invalid RFC3339 date: %v", err)
	}

	return core.V(float64(t.UnixMilli())), nil
}
