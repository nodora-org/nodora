package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func split() types.Func {
	return types.Func{
		Name:        "split",
		Description: "Splits a string into an array of substrings using the given delimiter.",
		Args: []types.ArgSpec{
			{
				Name:        "str",
				Description: "The string to split.",
				Type:        types.StringType,
				Required:    true,
			},
			{
				Name:        "delim",
				Description: "The delimiter to split on.",
				Type:        types.StringType,
				Required:    true,
			},
		},
		ReturnType: types.NewArrayType(types.StringType),
		Fn:         splitImpl,
		Pure:       true,
	}
}

func splitImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	delim := args[1]
	if str.IsUndefined() || delim.IsUndefined() {
		return core.U(), nil
	}
	strVal, ok := str.AsString()
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	delimVal, ok := delim.AsString()
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'delim' argument, got %v", delim.Type())
	}
	parts := strings.Split(strVal, delimVal)
	result := make([]core.Value, len(parts))
	for i, part := range parts {
		result[i] = core.V(part)
	}
	return core.V(result), nil
}
