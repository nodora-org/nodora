package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func contains() types.Func {
	return types.Func{
		Name:        "contains",
		Description: "Returns true if the string contains the given substring.",
		Args: []types.ArgSpec{
			{
				Name:        "str",
				Description: "The string to search in.",
				Type:        types.StringType,
				Required:    true,
			},
			{
				Name:        "substr",
				Description: "The substring to search for.",
				Type:        types.StringType,
				Required:    true,
			},
		},
		ReturnType: types.BoolType,
		Fn:         containsImpl,
	}
}

func containsImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	substr := args[1]
	if str.Undefined || substr.Undefined {
		return core.U(), nil
	}
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	substrVal, ok := substr.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'substr' argument, got %v", substr.Type())
	}
	return core.V(strings.Contains(strVal, substrVal)), nil
}
