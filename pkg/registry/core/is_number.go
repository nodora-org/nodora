package core

import (
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func isNumber() types.Func {
	return types.Func{
		Name:        "is_number",
		Description: "Returns true if the value is a number.",
		Args: []types.ArgSpec{{
			Name:        "value",
			Description: "The value to check.",
			Type:        types.AnyType,
			Required:    true,
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
