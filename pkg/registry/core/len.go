package core

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func length() types.Func {
	return types.Func{
		Name:        "len",
		Description: "Returns the length of a string, array, or object.",
		Args: []types.ArgSpec{{
			Name:        "value",
			Description: "The string, array, or object to measure.",
			Type:        types.NewUnionType(types.StringType, types.NewArrayType(types.AnyType), types.ObjectType),
			Required:    true,
		}},
		ReturnType: types.NumberType,
		Fn:         lengthImpl,
		Pure:       true,
	}
}

func lengthImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.IsUndefined() {
		return core.U(), nil
	}

	switch v := value.Raw.(type) {
	case string:
		return core.V(len(v)), nil
	case []core.Value:
		return core.V(len(v)), nil
	case core.ValueMap:
		return core.V(len(v)), nil
	default:
		return core.U(), fmt.Errorf(
			"expected %v for 'value' argument, got %v",
			types.NewUnionType(types.StringType, types.NewArrayType(types.AnyType), types.ObjectType),
			value.Type(),
		)
	}
}
