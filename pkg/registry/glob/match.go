package glob

import (
	"fmt"
	"path/filepath"

	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func match() types.Func {
	return types.Func{
		Name: "match",
		Args: []types.ArgSpec{
			{Name: "pattern", Type: types.StringType, Required: true},
			{Name: "subj", Type: types.StringType, Required: true},
		},
		ReturnType: types.BoolType,
		Fn:         matchImpl,
	}
}

func matchImpl(args []core.Value) (core.Value, error) {
	pattern := args[0]
	if pattern.Undefined {
		return core.U(), nil
	}
	subj := args[1]
	if subj.Undefined {
		return core.U(), nil
	}

	patternStr, ok := pattern.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'pattern' argument, got %v", pattern.Type())
	}
	subjStr, ok := subj.Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'subj' argument, got %v", subj.Type())
	}

	matched, err := filepath.Match(patternStr, subjStr)
	if err != nil {
		return core.U(), err
	}

	return core.V(matched), nil
}
