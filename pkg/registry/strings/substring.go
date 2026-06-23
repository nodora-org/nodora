package strings

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func substring() types.Func {
	return types.Func{
		Name:        "substring",
		Description: "Returns the portion of a string from start index to stop index.",
		Args: []types.ArgSpec{
			{
				Name:        "str",
				Description: "The input string.",
				Type:        types.StringType,
				Required:    true,
			},
			{
				Name:        "start",
				Description: "The start index (inclusive).",
				Type:        types.NumberType,
				Required:    true,
			},
			{
				Name:        "stop",
				Description: "The stop index (exclusive). Defaults to the string length.",
				Type:        types.NumberType,
				Required:    false,
			},
		},
		ReturnType: types.StringType,
		Fn:         substringImpl,
		Pure:       true,
	}
}

func substringImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	if str.Undefined {
		return core.U(), nil
	}
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}

	start := args[1]
	startVal, ok := core.ToInt(start.Raw)
	if !ok {
		return core.U(), fmt.Errorf("expected number for 'start' argument, got %v", start.Type())
	}

	startVal = clampIndex(startVal, 0, len(strVal))
	stopVal := len(strVal)

	if len(args) > 2 {
		stop := args[2]
		stopVal, ok = core.ToInt(stop.Raw)
		if !ok {
			return core.U(), fmt.Errorf("expected number for 'stop' argument, got %v", stop.Type())
		}
		stopVal = clampIndex(stopVal, 0, len(strVal))
	}

	if startVal >= stopVal {
		return core.V(""), nil
	}

	return core.V(strVal[startVal:stopVal]), nil
}

func clampIndex(x, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}
