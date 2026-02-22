package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func trimStart() types.Func {
	return types.Func{
		Name: "trim_start",
		Args: []types.ArgSpec{
			{
				Name:     "str",
				Type:     types.StringType,
				Required: true,
			},
			{
				Name:     "cutset",
				Type:     types.StringType,
				Required: true,
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
