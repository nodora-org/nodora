package core

import (
	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func isNumber() types.Func {
	return types.Func{
		Name: "is_number",
		Args: []types.ArgSpec{{
			Name:     "value",
			Type:     types.AnyType,
			Required: true,
		}},
		ReturnType: types.BoolType,
		Fn:         isNumberImpl,
	}
}

func isNumberImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	_, ok := core.ToFloat64(value)
	return core.V(ok), nil
}
