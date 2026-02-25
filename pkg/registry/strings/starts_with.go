package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func startsWith() types.Func {
	return types.Func{
		Name:        "starts_with",
		Description: "Returns true if the string starts with the given prefix.",
		Args: []types.ArgSpec{
			{
				Name:        "str",
				Description: "The string to check.",
				Type:        types.StringType,
				Required:    true,
			},
			{
				Name:        "prefix",
				Description: "The prefix to look for.",
				Type:        types.StringType,
				Required:    true,
			},
		},
		ReturnType: types.BoolType,
		Fn:         startsWithImpl,
	}
}

func startsWithImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	prefix := args[1]
	if str.Undefined || prefix.Undefined {
		return core.U(), nil
	}
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	prefixVal, ok := prefix.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'prefix' argument, got %v", prefix.Type())
	}
	return core.V(strings.HasPrefix(strVal, prefixVal)), nil
}
