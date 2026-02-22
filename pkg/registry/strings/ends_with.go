package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func endsWith() types.Func {
	return types.Func{
		Name: "ends_with",
		Args: []types.ArgSpec{
			{
				Name:     "str",
				Type:     types.StringType,
				Required: true,
			},
			{
				Name:     "suffix",
				Type:     types.StringType,
				Required: true,
			},
		},
		ReturnType: types.BoolType,
		Fn:         endsWithImpl,
	}
}

func endsWithImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	suffix := args[1]
	if str.Undefined || suffix.Undefined {
		return core.U(), nil
	}
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	suffixVal, ok := suffix.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'suffix' argument, got %v", suffix.Type())
	}
	return core.V(strings.HasSuffix(strVal, suffixVal)), nil
}
