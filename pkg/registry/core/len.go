package core

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/registry/types"
)

func length() types.Func {
	return types.Func{
		Name: "len",
		Args: []types.ArgSpec{{
			Name:     "value",
			Types:    []string{"string", "array", "object"},
			Required: true,
		}},
		ReturnType: "number",
		Fn:         lengthImpl,
	}
}

func lengthImpl(args []core.Value) (core.Value, error) {
	value := args[0]
	if value.Undefined {
		return core.U(), nil
	}
	switch v := value.Raw.(type) {
	case string:
		return core.V(len(v)), nil
	case []core.Value:
		return core.V(len(v)), nil
	case core.ValueMap:
		return core.V(len(v)), nil
	default:
		return core.U(), fmt.Errorf("type %v not supported", value.Type())
	}
}
