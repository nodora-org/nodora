package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func startsWith() types.Func {
	return types.Func{
		Name: "starts_with",
		Args: []types.ArgSpec{
			{
				Name:     "str",
				Types:    []string{"string"},
				Required: true,
			},
			{
				Name:     "prefix",
				Types:    []string{"string"},
				Required: true,
			},
		},
		ReturnType: "bool",
		Fn:         startsWithImpl,
	}
}

func startsWithImpl(args []core.Value) (core.Value, error) {
	strValue := args[0]
	prefixValue := args[1]
	if strValue.Undefined || prefixValue.Undefined {
		return core.U(), nil
	}
	str, ok := strValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for str", strValue.Type())
	}
	prefix, ok := prefixValue.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported for prefix", prefixValue.Type())
	}
	return core.V(strings.HasPrefix(str, prefix)), nil
}
