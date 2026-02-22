package objects

import (
	"fmt"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func isSubset() types.Func {
	return types.Func{
		Name: "is_subset",
		Args: []types.ArgSpec{
			{Name: "super", Type: types.ObjectType, Required: true},
			{Name: "sub", Type: types.ObjectType, Required: true},
		},
		ReturnType: types.ObjectType,
		Fn:         isSubsetImpl,
	}
}

func isSubsetImpl(args []core.Value) (core.Value, error) {
	super := args[0]
	if super.Undefined {
		return core.U(), nil
	}

	superVal, ok := super.Raw.(core.ValueMap)
	if !ok {
		return core.U(), fmt.Errorf("expected object for 'super' argument, got %v", super.Type())
	}

	sub := args[1]
	if sub.Undefined {
		return core.U(), nil
	}

	subVal, ok := sub.Raw.(core.ValueMap)
	if !ok {
		return core.U(), fmt.Errorf("expected object for 'sub' argument, got %v", sub.Type())
	}

	for key, subValue := range subVal {
		if subValue.Undefined {
			return core.V(false), nil
		}
		val, exists := superVal[key]
		if !exists {
			return core.V(false), nil
		}
		if !core.SafeEquals(val, subValue) {
			return core.V(false), nil
		}
	}

	return core.V(true), nil
}
