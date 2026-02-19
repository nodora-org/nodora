package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func endsWith() types.Func {
	return types.Func{
		Name: "ends_with",
		Args: []types.ArgSpec{
			{
				Name:     "str",
				Type:     types.StringType,
				Required: true,
			},
			{
				Name:     "suffix",
				Type:     types.StringType,
				Required: true,
			},
		},
		ReturnType: types.BoolType,
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
