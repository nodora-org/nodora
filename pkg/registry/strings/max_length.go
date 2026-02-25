package strings

import (
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func maxLength() types.Func {
	return types.Func{
		Name:        "max_length",
		Description: "Returns the length of the longest string in the array.",
		Args: []types.ArgSpec{{
			Name:        "arr",
			Description: "An array of strings.",
			Type:        types.NewArrayType(types.StringType),
			Required:    true,
		}},
		ReturnType: types.NumberType,
		Fn:         maxLengthImpl,
	}
}

func maxLengthImpl(args []core.Value) (core.Value, error) {
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

	var max int
	for i, item := range arrVal {
		val, ok := item.Raw.(string)
		if !ok || item.Undefined {
			return core.U(), nil
		}
		if i == 0 || len(val) > max {
			max = len(val)
		}
	}
	return core.V(max), nil
}
