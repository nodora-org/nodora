package objects

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func get() types.Func {
	t := types.NewTypeVar("T")
	return types.Func{
		Name:        "get",
		Description: "Gets a value from an object by key, returning a default if the key does not exist.",
		Args: []types.ArgSpec{
			{Name: "obj", Description: "The object to look up.", Type: types.ObjectType, Required: true},
			{Name: "key", Description: "The key to look up.", Type: types.StringType, Required: true},
			{Name: "default", Description: "The value to return if the key is not found.", Type: t, Required: true}},
		ReturnType: t,
		Fn:         getImpl,
		Pure:       true,
	}
}

func getImpl(args []core.Value) (core.Value, error) {
	obj := args[0]
	if obj.Undefined {
		return core.U(), nil
	}

	objVal, ok := obj.Raw.(core.ValueMap)
	if !ok {
		return core.U(), fmt.Errorf("expected object for 'obj' argument, got %v", obj.Type())
	}

	key := args[1]
	keyVal, ok := key.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'key' argument, got %v", key.Type())
	}

	defaultVal := args[2]
	val, exists := objVal[keyVal]
	if !exists {
		return defaultVal, nil
	}

	return val, nil
}
