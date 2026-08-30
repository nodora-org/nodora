package conv

import (
	"fmt"
	"strconv"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func atoi() types.Func {
	return types.Func{
		Name:        "atoi",
		Description: "Converts a string to an integer.",
		Args: []types.ArgSpec{{
			Name:        "str",
			Description: "The string to convert.",
			Type:        types.StringType,
			Required:    true,
		}},
		ReturnType: types.NumberType,
		Fn:         atoiImpl,
		Pure:       true,
	}
}

func atoiImpl(args []core.Value) (core.Value, error) {
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

	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return core.U(), fmt.Errorf("cannot parse %q as integer", strVal)
	}

	return core.V(intVal), nil
}
