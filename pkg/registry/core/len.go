package core

import (
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func length() types.Func {
	return types.Func{
		Name: "len",
		Args: []types.ArgSpec{{
			Name:     "value",
			Type:     types.NewArrayType(types.AnyType),
			Required: true,
		}},
		ReturnType: types.NumberType,
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
