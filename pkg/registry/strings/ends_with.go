package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func endsWith() types.Func {
	return types.Func{
		Name:        "ends_with",
		Description: "Returns true if the string ends with the given suffix.",
		Args: []types.ArgSpec{
			{
				Name:        "str",
				Description: "The string to check.",
				Type:        types.StringType,
				Required:    true,
			},
			{
				Name:        "suffix",
				Description: "The suffix to look for.",
				Type:        types.StringType,
				Required:    true,
			},
		},
		ReturnType: types.BoolType,
		Fn:         endsWithImpl,
		Pure:       true,
	}
}

func endsWithImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	suffix := args[1]
	if str.IsUndefined() || suffix.IsUndefined() {
		return core.U(), nil
	}
	strVal, ok := str.AsString()
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	suffixVal, ok := suffix.AsString()
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'suffix' argument, got %v", suffix.Type())
	}
	return core.V(strings.HasSuffix(strVal, suffixVal)), nil
}
