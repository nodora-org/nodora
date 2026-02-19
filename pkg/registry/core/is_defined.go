package core

import (
	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func isDefined() types.Func {
	return types.Func{
		Name:       "is_defined",
		Args:       []types.ArgSpec{{Name: "value", Type: types.AnyType, Required: true}},
		ReturnType: types.BoolType,
		Fn:         isDefinedImpl,
	}
}

func isDefinedImpl(args []core.Value) (core.Value, error) {
	return core.V(!args[0].Undefined), nil
}
