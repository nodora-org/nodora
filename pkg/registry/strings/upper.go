package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func upper() types.Func {
	return types.Func{
		Name:        "upper",
		Description: "Converts a string to uppercase.",
		Args: []types.ArgSpec{
			{
				Name:        "str",
				Description: "The string to convert.",
				Type:        types.StringType,
				Required:    true,
			},
		},
		ReturnType: types.StringType,
		Fn:         upperImpl,
	}
}

func upperImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	if str.Undefined {
		return core.U(), nil
	}
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	return core.V(strings.ToUpper(strVal)), nil
}
