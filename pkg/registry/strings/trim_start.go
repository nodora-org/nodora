package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func trimStart() types.Func {
	return types.Func{
		Name: "trim_start",
		Args: []types.ArgSpec{
			{
				Name:     "str",
				Types:    []string{"string"},
				Required: true,
			},
			{
				Name:     "cutset",
				Types:    []string{"string"},
				Required: true,
			},
		},
		ReturnType: "string",
		Fn:         trimStartImpl,
	}
}

func trimStartImpl(args []core.Value) (core.Value, error) {
	strValue := args[0]
	cutsetValue := args[1]
	if strValue.Undefined || cutsetValue.Undefined {
		return core.U(), nil
	}
	str, ok := strValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for str", strValue.Type())
	}
	cutset, ok := cutsetValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for cutset", cutsetValue.Type())
	}
	return core.V(strings.TrimLeft(str, cutset)), nil
}
