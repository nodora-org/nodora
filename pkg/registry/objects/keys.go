package objects

import (
	"fmt"
	"sort"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func keys() types.Func {
	return types.Func{
		Name:       "keys",
		Args:       []types.ArgSpec{{Name: "obj", Type: types.ObjectType, Required: true}},
		ReturnType: types.NewArrayType(types.StringType),
		Fn:         keysImpl,
	}
}

func keysImpl(args []core.Value) (core.Value, error) {
	obj := args[0]
	if obj.Undefined {
		return core.U(), nil
	}

	objVal, ok := obj.Raw.(core.ValueMap)
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
