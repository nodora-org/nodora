package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func endsWith() types.Func {
	return types.Func{
		Name: "ends_with",
		Args: []types.ArgSpec{
			{
				Name:     "str",
				Types:    []string{"string"},
				Required: true,
			},
			{
				Name:     "suffix",
				Types:    []string{"string"},
				Required: true,
			},
		},
		ReturnType: "bool",
		Fn:         endsWithImpl,
	}
}

func endsWithImpl(args []core.Value) (core.Value, error) {
	strValue := args[0]
	suffixValue := args[1]
	if strValue.Undefined || suffixValue.Undefined {
		return core.U(), nil
	}
	str, ok := strValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for str", strValue.Type())
	}
	suffix, ok := suffixValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for suffix", suffixValue.Type())
	}
	return core.V(strings.HasSuffix(str, suffix)), nil
}
