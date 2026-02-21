package objects

import (
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func remove() types.Func {
	return types.Func{
		Name: "remove",
		Args: []types.ArgSpec{
			{Name: "obj", Type: types.ObjectType, Required: true},
			{Name: "keys", Type: types.NewArrayType(types.StringType), Required: true},
		},
		ReturnType: types.ObjectType,
		Fn:         removeImpl,
	}
}

func removeImpl(args []core.Value) (core.Value, error) {
	obj := args[0]
	if obj.Undefined {
		return core.U(), nil
	}

	objVal, ok := obj.Raw.(core.ValueMap)
	if !ok {
		return core.U(), fmt.Errorf("expected object for 'obj' argument, got %T", obj.Type())
	}

	keys := args[1]
	keysVal, ok := keys.Raw.([]core.Value)
	if !ok {
		return core.U(), fmt.Errorf("expected array for 'keys' argument, got %T", obj.Type())
	}

	for _, k := range keysVal {
		keyVal, ok := k.Raw.(string)
		if !ok {
			return core.U(), fmt.Errorf("expected string in 'keys' array, got %T", obj.Type())
		}
		delete(objVal, keyVal)
	}

	return core.V(objVal), nil
}
