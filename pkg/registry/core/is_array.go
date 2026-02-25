package core

import (
	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func isArray() types.Func {
	return types.Func{
		Name:        "is_array",
		Description: "Returns true if the value is an array.",
		Args: []types.ArgSpec{{
			Name:        "value",
			Description: "The value to check.",
			Type:        types.AnyType,
			Required:    true,
		}},
		ReturnType: types.BoolType,
		Fn:         isArrayImpl,
	}
}

func isArrayImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	_, ok := value.Raw.([]core.Value)
	return core.V(ok), nil
}
