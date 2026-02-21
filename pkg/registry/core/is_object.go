package core

import (
	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func isObject() types.Func {
	return types.Func{
		Name: "is_object",
		Args: []types.ArgSpec{{
			Name:     "value",
			Type:     types.AnyType,
			Required: true,
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
