package objects

import (
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func get() types.Func {
	return types.Func{
		Name: "get",
		Args: []types.ArgSpec{
			{Name: "obj", Type: types.ObjectType, Required: true},
			{Name: "key", Type: types.StringType, Required: true},
			{Name: "default", Type: types.AnyType, Required: true}},
		ReturnType: types.AnyType,
		Fn:         getImpl,
	}
}

func getImpl(args []core.Value) (core.Value, error) {
	obj := args[0]
	if obj.Undefined {
		return core.U(), nil
	}

	objVal, ok := obj.Raw.(core.ValueMap)
	if !ok {
		return core.U(), fmt.Errorf("expected object for 'obj' argument, got %T", obj.Type())
	}

	key := args[1]
	keyVal, ok := key.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'key' argument, got %T", key.Type())
	}

	defaultVal := args[2]
	val, exists := objVal[keyVal]
	if !exists {
		return defaultVal, nil
	}

	return val, nil
}
