package core

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func isEmpty() types.Func {
	return types.Func{
		Name: "is_empty",
		Args: []types.ArgSpec{{
			Name:     "arr",
			Types:    []string{"array"},
			Required: true,
		}},
		ReturnType: "bool",
		Fn:         isEmptyImpl,
	}
}

func isEmptyImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	arr, ok := value.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf("type %v not supported", value.Type())
	}
	return core.V(len(arr) == 0), nil
}
