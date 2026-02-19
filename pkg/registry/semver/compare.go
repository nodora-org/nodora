package semver

import (
	"fmt"

	"golang.org/x/mod/semver"
	"nodora.org/nodora/internal/types"
	"nodora.org/nodora/pkg/core"
)

func compare() types.Func {
	return types.Func{
		Name: "compare",
		Args: []types.ArgSpec{
			{Name: "v", Type: types.StringType, Required: true},
			{Name: "w", Type: types.StringType, Required: true},
		},
		ReturnType: types.NumberType,
		Fn:         compareImpl,
	}
}

func compareImpl(args []core.Value) (core.Value, error) {
	v, ok := args[0].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid v type")
	}
	w, ok := args[1].Raw.(string)
	if !ok {
		return core.U(), fmt.Errorf("invalid w type")
	}
	return core.V(semver.Compare(v, w)), nil
}
