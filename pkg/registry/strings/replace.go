package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func replace() types.Func {
	return types.Func{
		Name: "replace",
		Args: []types.ArgSpec{
			{
				Name:     "str",
				Type:     types.StringType,
				Required: true,
			},
			{
				Name:     "old",
				Type:     types.StringType,
				Required: true,
			},
			{
				Name:     "new",
				Type:     types.StringType,
				Required: true,
			},
		},
		ReturnType: types.StringType,
		Fn:         replaceImpl,
	}
}

func replaceImpl(args []core.Value) (core.Value, error) {
	strValue := args[0]
	oldValue := args[1]
	newValue := args[2]
	if strValue.Undefined || oldValue.Undefined || newValue.Undefined {
		return core.U(), nil
	}
	str, ok := strValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for str", strValue.Type())
	}
	old, ok := oldValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for old", oldValue.Type())
	}
	new, ok := newValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for new", newValue.Type())
	}
	return core.V(strings.ReplaceAll(str, old, new)), nil
}
