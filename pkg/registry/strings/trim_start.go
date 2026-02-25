package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func trimStart() types.Func {
	return types.Func{
		Name:        "trim_start",
		Description: "Removes leading characters found in cutset from the string.",
		Args: []types.ArgSpec{
			{
				Name:        "str",
				Description: "The string to trim.",
				Type:        types.StringType,
				Required:    true,
			},
			{
				Name:        "cutset",
				Description: "The set of characters to remove from the start.",
				Type:        types.StringType,
				Required:    true,
			},
		},
		ReturnType: types.StringType,
		Fn:         trimStartImpl,
	}
}

func trimStartImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	cutset := args[1]
	if str.Undefined || cutset.Undefined {
		return core.U(), nil
	}
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	cutsetVal, ok := cutset.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'cutset' argument, got %v", cutset.Type())
	}
	return core.V(strings.TrimLeft(strVal, cutsetVal)), nil
}
