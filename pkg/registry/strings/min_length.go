package strings

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func minLength() types.Func {
	return types.Func{
		Name:        "min_length",
		Description: "Returns the length of the shortest string in the array.",
		Args: []types.ArgSpec{{
			Name:        "arr",
			Description: "An array of strings.",
			Type:        types.NewArrayType(types.StringType),
			Required:    true,
		}},
		ReturnType: types.NumberType,
		Fn:         minLengthImpl,
	}
}

func minLengthImpl(args []core.Value) (core.Value, error) {
	arr := args[0]
	if arr.Undefined {
		return core.U(), nil
	}

	arrVal, ok := arr.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf(
			"expected %v for 'arr' argument, got %v",
			types.NewArrayType(types.StringType),
			arr.Type(),
		)
	}

	if len(arrVal) == 0 {
		return core.U(), nil
	}

	var min int
	for i, item := range arrVal {
		val, ok := item.Raw.(string)
		if !ok || item.Undefined {
			return core.U(), nil
		}
		if i == 0 || len(val) < min {
			min = len(val)
		}
	}
	return core.V(min), nil
}
