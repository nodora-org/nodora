package strings

import (
	"fmt"
	"strings"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func upper() types.Func {
	return types.Func{
		Name: "upper",
		Args: []types.ArgSpec{
			{
				Name:     "str",
				Types:    []string{"string"},
				Required: true,
			},
		},
		ReturnType: "string",
		Fn:         upperImpl,
	}
}

func upperImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	str, ok := value.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported", value.Type())
	}
	return core.V(strings.ToUpper(str)), nil
}
