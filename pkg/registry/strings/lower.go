package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func lower() types.Func {
	return types.Func{
		Name:        "lower",
		Description: "Converts a string to lowercase.",
		Args: []types.ArgSpec{
			{
				Name:        "str",
				Description: "The string to convert.",
				Type:        types.StringType,
				Required:    true,
			},
		},
		ReturnType: types.StringType,
		Fn:         lowerImpl,
	}
}

func lowerImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	if str.Undefined {
		return core.U(), nil
	}
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	return core.V(strings.ToLower(strVal)), nil
}
