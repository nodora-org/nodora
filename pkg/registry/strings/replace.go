package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func replace() types.Func {
	return types.Func{
		Name:        "replace",
		Description: "Replaces all occurrences of a substring with a replacement string.",
		Args: []types.ArgSpec{
			{
				Name:        "str",
				Description: "The input string.",
				Type:        types.StringType,
				Required:    true,
			},
			{
				Name:        "old",
				Description: "The substring to replace.",
				Type:        types.StringType,
				Required:    true,
			},
			{
				Name:        "new",
				Description: "The replacement string.",
				Type:        types.StringType,
				Required:    true,
			},
		},
		ReturnType: types.StringType,
		Fn:         replaceImpl,
		Pure:       true,
	}
}

func replaceImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	old := args[1]
	new := args[2]
	if str.Undefined || old.Undefined || new.Undefined {
		return core.U(), nil
	}
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	oldVal, ok := old.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'old' argument, got %v", old.Type())
	}
	newVal, ok := new.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'new' argument, got %v", new.Type())
	}
	return core.V(strings.ReplaceAll(strVal, oldVal, newVal)), nil
}
