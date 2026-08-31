package objects

import (
	"fmt"
	"sort"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func keys() types.Func {
	return types.Func{
		Name:        "keys",
		Description: "Returns a sorted array of an object's keys.",
		Args:        []types.ArgSpec{{Name: "obj", Description: "The object to extract keys from.", Type: types.ObjectType, Required: true}},
		ReturnType:  types.NewArrayType(types.StringType),
		Fn:          keysImpl,
		Pure:        true,
	}
}

func keysImpl(args []core.Value) (core.Value, error) {
	obj := args[0]
	if obj.IsUndefined() {
		return core.U(), nil
	}

	objVal, ok := obj.AsObject()
	if !ok {
		return core.U(), fmt.Errorf("expected object for 'obj' argument, got %v", obj.Type())
	}

	keys := make([]string, 0, len(objVal))
	for k := range objVal {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	return core.V(keys), nil
}
