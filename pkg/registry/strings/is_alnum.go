package strings

import (
	"fmt"
	"unicode"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func isAlphanumeric() types.Func {
	return types.Func{
		Name: "is_alnum",
		Args: []types.ArgSpec{{
			Name:     "str",
			Type:     types.StringType,
			Required: true,
		}},
		ReturnType: types.BoolType,
		Fn:         isAlphanumericImpl,
	}
}

func isAlphanumericImpl(args []core.Value) (core.Value, error) {
	str := args[0]
	if str.Undefined {
		return core.U(), nil
	}
	strVal, ok := str.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'str' argument, got %v", str.Type())
	}
	for _, r := range strVal {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return core.V(false), nil
		}
	}
	return core.V(len(strVal) > 0), nil
}
