package objects

import (
	"fmt"
	"sort"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func values() types.Func {
	return types.Func{
		Name:       "values",
		Args:       []types.ArgSpec{{Name: "obj", Type: types.ObjectType, Required: true}},
		ReturnType: types.NewArrayType(types.AnyType),
		Fn:         valuesImpl,
	}
}

func valuesImpl(args []core.Value) (core.Value, error) {
	obj := args[0]
	if obj.Undefined {
		return core.U(), nil
	}

	objVal, ok := obj.Raw.(core.ValueMap)
	if !ok {
		return core.U(), fmt.Errorf("expected object for 'obj' argument, got %T", obj.Type())
	}

	keys := make([]string, 0, len(objVal))
	for k := range objVal {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	values := make([]core.Value, 0, len(objVal))
	for _, k := range keys {
		values = append(values, objVal[k])
	}

	return core.V(values), nil
}
