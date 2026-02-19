package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func contains() types.Func {
	return types.Func{
		Name: "contains",
		Args: []types.ArgSpec{
			{
				Name:     "str",
				Type:     types.StringType,
				Required: true,
			},
			{
				Name:     "substr",
				Type:     types.StringType,
				Required: true,
			},
		},
		ReturnType: types.BoolType,
		Fn:         containsImpl,
	}
}

func containsImpl(args []core.Value) (core.Value, error) {
	strValue := args[0]
	substrValue := args[1]
	if strValue.Undefined || substrValue.Undefined {
		return core.U(), nil
	}
	str, ok := strValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for str", strValue.Type())
	}
	substr, ok := substrValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for substr", substrValue.Type())
	}
	return core.V(strings.Contains(str, substr)), nil
}
