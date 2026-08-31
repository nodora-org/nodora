package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
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
		Pure:       true,
	}
}

func trimStartImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	cutset := args[1]
	if str.IsUndefined() || cutset.IsUndefined() {
		return core.U(), nil
	}
	strVal, ok := str.AsString()
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	cutsetVal, ok := cutset.AsString()
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'cutset' argument, got %v", cutset.Type())
	}
	return core.V(strings.TrimLeft(strVal, cutsetVal)), nil
}
