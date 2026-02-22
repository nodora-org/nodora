package conv

import (
	"fmt"
	"strconv"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func atoi() types.Func {
	return types.Func{
		Name: "atoi",
		Args: []types.ArgSpec{{
			Name:     "str",
			Type:     types.StringType,
			Required: true,
		}},
		ReturnType: types.NumberType,
		Fn:         atoiImpl,
	}
}

func atoiImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	if str.Undefined {
		return core.U(), nil
	}
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v",
			str.Type(),
		)
	}

	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return core.U(), err
	}

	return core.V(intVal), nil
}
