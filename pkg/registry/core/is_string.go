package core

import (
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func isString() types.Func {
	return types.Func{
		Name:        "is_string",
		Description: "Returns true if the value is a string.",
		Args: []types.ArgSpec{{
			Name:        "value",
			Description: "The value to check.",
			Type:        types.AnyType,
			Required:    true,
		}},
		ReturnType: types.BoolType,
		Fn:         isStringImpl,
		Pure:       true,
	}
}

func isStringImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.IsUndefined() {
		return core.U(), nil
	}
	_, ok := value.AsString()
	return core.V(ok), nil
}
