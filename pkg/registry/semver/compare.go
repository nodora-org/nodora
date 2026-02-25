package semver

import (
	"fmt"

	"golang.org/x/mod/semver"
	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func compare() types.Func {
	return types.Func{
		Name:        "compare",
		Description: "Compares two semantic versions. Returns -1, 0, or 1.",
		Args: []types.ArgSpec{
			{Name: "v", Description: "The first version string.", Type: types.StringType, Required: true},
			{Name: "w", Description: "The second version string.", Type: types.StringType, Required: true},
		},
		ReturnType: types.NumberType,
		Fn:         compareImpl,
	}
}

func compareImpl(args []core.Value) (core.Value, error) {
	v := args[0]
	vv, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'v' argument, got %v", v.Type())
	}
	w := args[1]
	ww, ok := args[1].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("expected string for 'w' argument, got %v", w.Type())
	}
	return core.V(semver.Compare(vv, ww)), nil
}
