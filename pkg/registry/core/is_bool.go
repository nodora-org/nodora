package core

import (
	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func isBool() types.Func {
	return types.Func{
		Name: "is_bool",
		Args: []types.ArgSpec{{
			Name:     "value",
			Type:     types.AnyType,
			Required: true,
		}},
		ReturnType: types.BoolType,
		Fn:         isBoolImpl,
	}
}

func isBoolImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	_, ok := value.Raw.(bool)
	return core.V(ok), nil
}
