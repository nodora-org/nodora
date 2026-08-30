package conv

import (
	"fmt"
	"strconv"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func atof() types.Func {
	return types.Func{
		Name:        "atof",
		Description: "Converts a string to a floating-point number.",
		Args: []types.ArgSpec{{
			Name:        "str",
			Description: "The string to convert.",
			Type:        types.StringType,
			Required:    true,
		}},
		ReturnType: types.NumberType,
		Fn:         atofImpl,
		Pure:       true,
	}
}

func atofImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	if str.IsUndefined() {
		return core.U(), nil
	}
	strVal, ok := str.AsString()
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v",
			str.Type(),
		)
	}

	floatVal, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		return core.U(), fmt.Errorf("cannot parse %q as float", strVal)
	}

	return core.V(floatVal), nil
}
