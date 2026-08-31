package strings

import (
	"fmt"
	"unicode"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func isAlpha() types.Func {
	return types.Func{
		Name:        "is_alpha",
		Description: "Returns true if the string is non-empty and contains only alphabetic characters.",
		Args: []types.ArgSpec{{
			Name:        "str",
			Description: "The string to check.",
			Type:        types.StringType,
			Required:    true,
		}},
		ReturnType: types.BoolType,
		Fn:         isAlphaImpl,
		Pure:       true,
	}
}

func isAlphaImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	if str.IsUndefined() {
		return core.U(), nil
	}
	strVal, ok := str.AsString()
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	for _, r := range strVal {
		if !unicode.IsLetter(r) {
			return core.V(false), nil
		}
	}
	return core.V(len(strVal) > 0), nil
}
