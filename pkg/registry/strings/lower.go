package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func lower() types.Func {
	return types.Func{
		Name: "lower",
		Args: []types.ArgSpec{
			{
				Name:     "str",
				Types:    []string{"string"},
				Required: true,
			},
		},
		ReturnType: "string",
		Fn:         lowerImpl,
	}
}

func lowerImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	str, ok := value.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported", value.Type())
	}
	return core.V(strings.ToLower(str)), nil
}
