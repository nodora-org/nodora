package strings

import (
	"fmt"
	"unicode"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func isAlpha() types.Func {
	return types.Func{
		Name: "is_alpha",
		Args: []types.ArgSpec{{
			Name:     "str",
			Types:    []string{"string"},
			Required: true,
		}},
		ReturnType: "bool",
		Fn:         isAlphaImpl,
	}
}

func isAlphaImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	str, ok := value.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported", value.Type())
	}
	for _, r := range str {
		if !unicode.IsLetter(r) {
			return core.V(false), nil
		}
	}
	return core.V(len(str) > 0), nil
}
