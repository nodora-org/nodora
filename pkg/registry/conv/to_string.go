package conv

import (
	"fmt"
	"strconv"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func toString() types.Func {
	return types.Func{
		Name:        "to_string",
		Description: "Converts a number, bool, or string to its string representation.",
		Args: []types.ArgSpec{{
			Name:        "x",
			Description: "The value to convert.",
			Type:        types.NewUnionType(types.NumberType, types.BoolType, types.StringType),
			Required:    true,
		}},
		ReturnType: types.StringType,
		Fn:         toStringImpl,
		Pure:       true,
	}
}

func toStringImpl(args []core.Value) (core.Value, error) {
	x := args[0]
	if x.Undefined {
		return core.U(), nil
	}
	switch v := x.Raw.(type) {
	case string:
		return core.V(v), nil
	case bool:
		return core.V(strconv.FormatBool(v)), nil
	case float64:
		return core.V(strconv.FormatFloat(v, 'f', -1, 64)), nil
	default:
		return core.U(), fmt.Errorf(
			"expected %v for 'x' argument, got %v",
			types.NewUnionType(types.NumberType, types.BoolType, types.StringType),
			x.Type(),
		)
	}
}
