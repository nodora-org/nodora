package core

import (
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func isDefined() types.Func {
	return types.Func{
		Name:       "is_defined",
		Args:       []types.ArgSpec{{Name: "value", Types: nil, Required: true}},
		ReturnType: "bool",
		Fn:         isDefinedImpl,
	}
}

func isDefinedImpl(args []core.Value) (core.Value, error) {
	return core.V(!args[0].Undefined), nil
}
