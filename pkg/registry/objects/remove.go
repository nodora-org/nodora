package objects

import (
	"fmt"
	"maps"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func remove() types.Func {
	return types.Func{
		Name:        "remove",
		Description: "Removes the specified keys from an object.",
		Args: []types.ArgSpec{
			{Name: "obj", Description: "The object to remove keys from.", Type: types.ObjectType, Required: true},
			{Name: "keys", Description: "An array of key names to remove.", Type: types.NewArrayType(types.StringType), Required: true},
		},
		ReturnType: types.ObjectType,
		Fn:         removeImpl,
		Pure:       true,
	}
}

func removeImpl(args []core.Value) (core.Value, error) {
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
		return core.U(), fmt.Errorf("expected array for 'keys' argument, got %v", keys.Type())
	}

	result := maps.Clone(objVal)
	for _, k := range keysVal {
		keyVal, ok := k.Raw.(string)
		if !ok {
			return core.U(), fmt.Errorf("expected string in 'keys' array, got %v", k.Type())
		}
		delete(result, keyVal)
	}

	return core.V(result), nil
}
