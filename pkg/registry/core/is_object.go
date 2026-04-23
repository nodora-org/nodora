package core

import (
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func isObject() types.Func {
	return types.Func{
		Name:        "is_object",
		Description: "Returns true if the value is an object.",
		Args: []types.ArgSpec{{
			Name:        "value",
			Description: "The value to check.",
			Type:        types.AnyType,
			Required:    true,
		}},
		ReturnType: types.BoolType,
		Fn:         isObjectImpl,
	}
}

func isObjectImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	_, ok := value.Raw.(core.ValueMap)
	return core.V(ok), nil
}
