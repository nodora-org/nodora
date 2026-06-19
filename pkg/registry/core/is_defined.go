package core

import (
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func isDefined() types.Func {
	return types.Func{
		Name:        "is_defined",
		Description: "Returns true if the value is defined (not undefined).",
		Args:        []types.ArgSpec{{Name: "value", Description: "The value to check.", Type: types.AnyType, Required: true}},
		ReturnType:       types.BoolType,
		Fn:               isDefinedImpl,
		Pure:             true,
		AcceptsUndefined: true,
	}
}

func isDefinedImpl(args []core.Value) (core.Value, error) {
	return core.V(!args[0].Undefined), nil
}
