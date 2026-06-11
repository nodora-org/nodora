package objects

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func filter() types.Func {
	return types.Func{
		Name:        "filter",
		Description: "Returns a new object containing only the specified keys.",
		Args: []types.ArgSpec{
			{Name: "obj", Description: "The source object.", Type: types.ObjectType, Required: true},
			{Name: "keys", Description: "An array of key names to keep.", Type: types.NewArrayType(types.StringType), Required: true},
		},
		ReturnType: types.ObjectType,
		Fn:         filterImpl,
		Pure:       true,
	}
}

func filterImpl(args []core.Value) (core.Value, error) {
	obj := args[0]
	if obj.Undefined {
		return core.U(), nil
	}

	objVal, ok := obj.Raw.(core.ValueMap)
	if !ok {
		return core.U(), fmt.Errorf("expected object for 'obj' argument, got %v", obj.Type())
	}

	keys := args[1]
	keysVal, ok := keys.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf("expected array for 'keys' argument, got %v", obj.Type())
	}

	result := make(core.ValueMap)
	for _, k := range keysVal {
		keyVal, ok := k.Raw.(string)
		if !ok {
			return core.U(), fmt.Errorf("expected string in 'keys' array, got %v", k.Raw)
		}
		if val, exists := objVal[keyVal]; exists {
			result[keyVal] = val
		}
	}

	return core.V(result), nil
}
