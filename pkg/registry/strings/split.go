package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func split() types.Func {
	return types.Func{
		Name: "split",
		Args: []types.ArgSpec{
			{
				Name:     "str",
				Type:     types.StringType,
				Required: true,
			},
			{
				Name:     "delim",
				Type:     types.StringType,
				Required: true,
			},
		},
		ReturnType: types.NewArrayType(types.StringType),
		Fn:         splitImpl,
	}
}

func splitImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	delim := args[1]
	if str.Undefined || delim.Undefined {
		return core.U(), nil
	}
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	delimVal, ok := delim.Raw.(string)
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
