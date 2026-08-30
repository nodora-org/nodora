package core

import (
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func isBool() types.Func {
	return types.Func{
		Name:        "is_bool",
		Description: "Returns true if the value is a boolean.",
		Args: []types.ArgSpec{{
			Name:        "value",
			Description: "The value to check.",
			Type:        types.AnyType,
			Required:    true,
		}},
		ReturnType: types.BoolType,
		Fn:         isBoolImpl,
		Pure:       true,
	}
}

func isBoolImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.IsUndefined() {
		return core.U(), nil
	}
	_, ok := value.AsBool()
	return core.V(ok), nil
}
