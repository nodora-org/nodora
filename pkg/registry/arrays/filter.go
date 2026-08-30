package arrays

import (
	"fmt"

	"nodora.org/nodora/pkg/core"
	"nodora.org/nodora/pkg/types"
)

func filter() types.Func {
	t := types.NewTypeVar("T")
	return types.Func{
		Name:        "filter",
		Description: "Returns a new array containing only the elements that satisfy the predicate.",
		Args: []types.ArgSpec{
			{
				Name:        "arr",
				Description: "The array to filter.",
				Type:        types.NewArrayType(t),
				Required:    true,
			},
			{
				Name:        "fx",
				Description: "A predicate function applied to each element.",
				Type:        types.NewLambdaType([]types.Type{t}, types.BoolType),
				Required:    true,
			},
		},
		ReturnType: types.NewArrayType(t),
		Fn:         filterImpl,
		Pure:       true,
	}
}

func filterImpl(args []core.Value) (core.Value, error) {
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

	fx := args[1]
	if fx.IsUndefined() {
		return core.U(), nil
	}
	fn, ok := fx.AsLambda()
	if !ok {
		return core.U(), fmt.Errorf(
			"expected %v for 'fx' argument, got %v",
			types.NewLambdaType([]types.Type{types.AnyType}, types.BoolType),
			fx.Type(),
		)
	}

	result := []core.Value{}
	for _, v := range arrVal {
		r, err := (*fn).Fn([]core.Value{v})
		if err != nil {
			return core.U(), err
		}
		if r.IsUndefined() {
			return core.U(), nil
		}
		keep, ok := r.AsBool()
		if !ok {
			return core.U(), fmt.Errorf("expected a bool value, got %v", r.Type())
		}
		if keep {
			result = append(result, v)
		}
	}

	return core.V(result), nil
}
