package arrays

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func unique() types.Func {
	t := types.NewTypeVar("T")
	return types.Func{
		Name:        "unique",
		Description: "Returns a new array with duplicate elements removed, preserving first-seen order.",
		Args: []types.ArgSpec{
			{
				Name:        "arr",
				Description: "The array to deduplicate.",
				Type:        types.NewArrayType(t),
				Required:    true,
			},
		},
		ReturnType: types.NewArrayType(t),
		Fn:         uniqueImpl,
		Pure:       true,
	}
}

func uniqueImpl(args []core.Value) (core.Value, error) {
	arr := args[0]
	if arr.IsUndefined() {
		return core.U(), nil
	}
	arrVal, ok := arr.AsArray()
	if !ok {
		return core.U(), fmt.Errorf(
			"expected %v for 'arr' argument, got %v",
			types.NewArrayType(types.AnyType),
			arr.Type(),
		)
	}

	result := []core.Value{}
	for _, v := range arrVal {
		seen := false
		for _, kept := range result {
			if core.SafeEquals(v, kept) {
				seen = true
				break
			}
		}
		if !seen {
			result = append(result, v)
		}
	}

	return core.V(result), nil
}
