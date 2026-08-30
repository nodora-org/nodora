package core

import (
	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func fallback() types.Func {
	t := types.NewTypeVar("T")
	return types.Func{
		Name:        "fallback",
		Description: "Returns the value if defined, otherwise returns the fallback value.",
		Args: []types.ArgSpec{
			{
				Name:        "value",
				Description: "The value to return if defined.",
				Type:        t,
				Required:    true,
			},
			{
				Name:        "fb",
				Description: "The value to return if `value` is undefined.",
				Type:        t,
				Required:    true,
			}},
		ReturnType:       t,
		Fn:               fallbackImpl,
		Pure:             true,
		AcceptsUndefined: true,
	}
}

func fallbackImpl(args []core.Value) (core.Value, error) {
	if args[0].IsUndefined() {
		return args[1], nil
	}
	return args[0], nil
}
