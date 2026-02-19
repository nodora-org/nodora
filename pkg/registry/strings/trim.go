package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func trim() types.Func {
	return types.Func{
		Name: "trim",
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
		Fn:         trimImpl,
	}
}

func trimImpl(args []core.Value) (core.Value, error) {
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
	return core.V(strings.Trim(str, cutset)), nil
}
