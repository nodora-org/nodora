package objects

import (
	"fmt"
	"maps"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func union() types.Func {
	return types.Func{
		Name:        "union",
		Description: "Merges two objects. Keys from obj2 overwrite keys from obj1.",
		Args: []types.ArgSpec{
			{Name: "obj1", Description: "The base object.", Type: types.ObjectType, Required: true},
			{Name: "obj2", Description: "The object to merge in (takes precedence).", Type: types.ObjectType, Required: true},
		},
		ReturnType: types.ObjectType,
		Fn:         unionImpl,
	}
}

func unionImpl(args []core.Value) (core.Value, error) {
	obj1 := args[0]
	if obj1.Undefined {
		return core.U(), nil
	}

	obj1Val, ok := obj1.Raw.(core.ValueMap)
	if !ok {
		return core.U(), fmt.Errorf("expected object for 'obj1' argument, got %v", obj1.Type())
	}

	obj2 := args[1]
	if obj2.Undefined {
		return core.U(), nil
	}

	obj2Val, ok := obj2.Raw.(core.ValueMap)
	if !ok {
		return core.U(), fmt.Errorf("expected object for 'obj2' argument, got %v", obj2.Type())
	}

	maps.Copy(obj1Val, obj2Val)

	return core.V(obj1Val), nil
}
