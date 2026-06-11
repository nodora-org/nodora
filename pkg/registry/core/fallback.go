package core

import (
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func fallback() types.Func {
	return types.Func{
		Name:        "fallback",
		Description: "Returns the value if defined, otherwise returns the fallback value.",
		Args: []types.ArgSpec{
			{
				Name:        "value",
				Description: "The value to return if defined.",
				Type:        types.AnyType,
				Required:    true,
			},
			{
				Name:        "fb",
				Description: "The value to return if `value` is undefined.",
				Type:        types.AnyType,
				Required:    true,
			}},
		ReturnType: types.AnyType,
		Fn:         fallbackImpl,
		Pure:       true,
	}
}

func fallbackImpl(args []core.Value) (core.Value, error) {
	if args[0].Undefined {
		return args[1], nil
	}
	return args[0], nil
}
