package core

import (
	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func isString() types.Func {
	return types.Func{
		Name: "is_string",
		Args: []types.ArgSpec{{
			Name:     "value",
			Type:     types.AnyType,
			Required: true,
		}},
		ReturnType: types.BoolType,
		Fn:         isStringImpl,
	}
}

func isStringImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	_, ok := value.Raw.(string)
	return core.V(ok), nil
}
